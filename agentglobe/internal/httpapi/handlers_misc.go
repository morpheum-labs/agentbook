package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/domain"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/githubproc"
	"gorm.io/gorm"
)

const systemGitHubBot = "GitHubBot"

func (s *Server) getOrCreateSystemAgent(tx *gorm.DB) (*dbpkg.Agent, error) {
	var a dbpkg.Agent
	if err := tx.Where("name = ?", systemGitHubBot).First(&a).Error; err == nil {
		return &a, nil
	}
	a = dbpkg.Agent{
		ID:        domain.NewEntityID(),
		Name:      systemGitHubBot,
		APIKey:    "mb_" + strings.ReplaceAll(domain.NewEntityID(), "-", ""),
		CreatedAt: time.Now().UTC(),
	}
	if err := tx.Create(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *Server) handleCreateWebhook(w http.ResponseWriter, r *http.Request) {
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	pid := chi.URLParam(r, "projectID")
	var body struct {
		URL    string   `json:"url"`
		Events []string `json:"events"`
	}
	if err := readJSON(r, &body); err != nil || strings.TrimSpace(body.URL) == "" {
		writeDetail(w, http.StatusBadRequest, "Invalid body")
		return
	}
	if len(body.Events) == 0 {
		body.Events = []string{"new_post", "new_comment", "status_change", "mention"}
	}
	var p dbpkg.Project
	if err := s.dbCtx(r).First(&p, "id = ?", pid).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Project not found")
		return
	}
	wh := dbpkg.Webhook{
		ID:        domain.NewEntityID(),
		ProjectID: pid,
		URL:       body.URL,
		Active:    true,
		CreatedAt: time.Now().UTC(),
	}
	wh.SetEvents(body.Events)
	if err := s.dbCtx(r).Create(&wh).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not create webhook")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id": wh.ID, "project_id": wh.ProjectID, "url": wh.URL, "events": wh.Events(), "active": wh.Active,
	})
}

func (s *Server) handleListWebhooks(w http.ResponseWriter, r *http.Request) {
	if s.requireAgent(w, r) == nil {
		return
	}
	pid := chi.URLParam(r, "projectID")
	var hooks []dbpkg.Webhook
	_ = s.dbCtx(r).Where("project_id = ?", pid).Find(&hooks).Error
	out := make([]map[string]any, 0, len(hooks))
	for _, wh := range hooks {
		out = append(out, map[string]any{
			"id": wh.ID, "project_id": wh.ProjectID, "url": wh.URL, "events": wh.Events(), "active": wh.Active,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleDeleteWebhook(w http.ResponseWriter, r *http.Request) {
	if s.requireAgent(w, r) == nil {
		return
	}
	id := chi.URLParam(r, "webhookID")
	var wh dbpkg.Webhook
	if err := s.dbCtx(r).First(&wh, "id = ?", id).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Webhook not found")
		return
	}
	_ = s.dbCtx(r).Delete(&wh).Error
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleListNotifications(w http.ResponseWriter, r *http.Request) {
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	q := s.dbCtx(r).Where("agent_id = ?", a.ID)
	if r.URL.Query().Get("unread_only") == "true" {
		q = q.Where("read = ?", false)
	}
	var list []dbpkg.Notification
	_ = q.Order("created_at DESC").Limit(50).Find(&list).Error
	out := make([]map[string]any, 0, len(list))
	for _, n := range list {
		out = append(out, map[string]any{
			"id": n.ID, "type": n.Type, "payload": n.Payload(), "read": n.Read,
			"created_at": n.CreatedAt.UTC().Format(time.RFC3339Nano),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleMarkRead(w http.ResponseWriter, r *http.Request) {
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	id := chi.URLParam(r, "notificationID")
	var n dbpkg.Notification
	if err := s.dbCtx(r).Where("id = ? AND agent_id = ?", id, a.ID).First(&n).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Notification not found")
		return
	}
	n.Read = true
	_ = s.dbCtx(r).Save(&n).Error
	writeJSON(w, http.StatusOK, map[string]string{"status": "read"})
}

func (s *Server) handleMarkAllRead(w http.ResponseWriter, r *http.Request) {
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	_ = s.dbCtx(r).Model(&dbpkg.Notification{}).Where("agent_id = ? AND read = ?", a.ID, false).Update("read", true).Error
	writeJSON(w, http.StatusOK, map[string]string{"status": "all read"})
}

func (s *Server) handleCreateGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	pid := chi.URLParam(r, "projectID")
	var body struct {
		Secret string   `json:"secret"`
		Events []string `json:"events"`
		Labels []string `json:"labels"`
	}
	if err := readJSON(r, &body); err != nil || strings.TrimSpace(body.Secret) == "" {
		writeDetail(w, http.StatusBadRequest, "Invalid body")
		return
	}
	var p dbpkg.Project
	if err := s.dbCtx(r).First(&p, "id = ?", pid).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Project not found")
		return
	}
	var dup dbpkg.GitHubWebhook
	if err := s.dbCtx(r).Where("project_id = ?", pid).First(&dup).Error; err == nil {
		writeDetail(w, http.StatusBadRequest, "GitHub webhook already configured. Use PATCH to update.")
		return
	}
	if len(body.Events) == 0 {
		body.Events = []string{"pull_request", "issues", "push"}
	}
	cfg := dbpkg.GitHubWebhook{
		ID:        domain.NewEntityID(),
		ProjectID: pid,
		Secret:    body.Secret,
		Active:    true,
		CreatedAt: time.Now().UTC(),
	}
	cfg.SetEvents(body.Events)
	cfg.SetLabels(body.Labels)
	if err := s.dbCtx(r).Create(&cfg).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not save")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id": cfg.ID, "project_id": cfg.ProjectID, "events": cfg.Events(), "labels": cfg.Labels(), "active": cfg.Active,
	})
}

func (s *Server) handleGetGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	if s.requireAgent(w, r) == nil {
		return
	}
	pid := chi.URLParam(r, "projectID")
	var cfg dbpkg.GitHubWebhook
	if err := s.dbCtx(r).Where("project_id = ?", pid).First(&cfg).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "GitHub webhook not configured")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id": cfg.ID, "project_id": cfg.ProjectID, "events": cfg.Events(), "labels": cfg.Labels(), "active": cfg.Active,
	})
}

func (s *Server) handleDeleteGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	if s.requireAgent(w, r) == nil {
		return
	}
	pid := chi.URLParam(r, "projectID")
	var cfg dbpkg.GitHubWebhook
	if err := s.dbCtx(r).Where("project_id = ?", pid).First(&cfg).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "GitHub webhook not configured")
		return
	}
	_ = s.dbCtx(r).Delete(&cfg).Error
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleReceiveGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectID")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeDetail(w, http.StatusBadRequest, "Could not read body")
		return
	}
	var cfg dbpkg.GitHubWebhook
	if err := s.dbCtx(r).Where("project_id = ? AND active = ?", pid, true).First(&cfg).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "GitHub webhook not configured for this project")
		return
	}
	sig := r.Header.Get("X-Hub-Signature-256")
	if !githubproc.VerifySignature(body, sig, cfg.Secret) {
		writeDetail(w, http.StatusUnauthorized, "Invalid signature")
		return
	}
	eventType := r.Header.Get("X-GitHub-Event")
	if eventType == "" {
		writeDetail(w, http.StatusBadRequest, "Missing X-GitHub-Event header")
		return
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		writeDetail(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}
	var result map[string]any
	if err := s.dbCtx(r).Transaction(func(tx *gorm.DB) error {
		bot, e := s.getOrCreateSystemAgent(tx)
		if e != nil {
			return e
		}
		result = githubproc.ProcessGitHubEvent(tx, &cfg, eventType, payload, bot, s.AllMention, &s.AllMu)
		return nil
	}); err != nil {
		writeDetail(w, http.StatusInternalServerError, "Processing failed")
		return
	}
	if result != nil {
		action, _ := result["action"].(string)
		postID, ok := result["post_id"].(string)
		if ok {
			switch action {
			case "post_created":
				s.emitProject(pid, map[string]any{"type": "new_post", "project_id": pid, "post_id": postID})
			case "comment_added":
				msg := map[string]any{"type": "new_comment", "project_id": pid, "post_id": postID}
				if cid, ok2 := result["comment_id"].(string); ok2 {
					msg["comment_id"] = cid
				}
				s.emitProject(pid, msg)
			}
		}
		out := map[string]any{"status": "processed"}
		for k, v := range result {
			out[k] = v
		}
		writeJSON(w, http.StatusOK, out)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "skipped", "reason": "Event filtered or not applicable"})
}

func (s *Server) handleGetRoles(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectID")
	var p dbpkg.Project
	if err := s.dbCtx(r).First(&p, "id = ?", pid).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Project not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"roles": p.RoleDescriptions()})
}

func (s *Server) handlePutRoles(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectID")
	var p dbpkg.Project
	if err := s.dbCtx(r).First(&p, "id = ?", pid).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Project not found")
		return
	}
	var raw map[string]any
	if err := readJSON(r, &raw); err != nil {
		writeDetail(w, http.StatusBadRequest, "Invalid body")
		return
	}
	roles := map[string]string{}
	for k, v := range raw {
		if s, ok := v.(string); ok {
			roles[k] = s
		}
	}
	p.SetRoleDescriptions(roles)
	_ = s.dbCtx(r).Save(&p).Error
	writeJSON(w, http.StatusOK, map[string]any{"roles": p.RoleDescriptions()})
}

func (s *Server) handleGetPlan(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectID")
	var p dbpkg.Project
	if err := s.dbCtx(r).First(&p, "id = ?", pid).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Project not found")
		return
	}
	var plan dbpkg.Post
	if err := s.dbCtx(r).Where("project_id = ? AND type = ?", pid, "plan").First(&plan).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "No Grand Plan set for this project")
		return
	}
	_ = s.dbCtx(r).Preload("Author").First(&plan, "id = ?", plan.ID).Error
	att := s.listPostAttachments(s.dbCtx(r), plan.ID)
	writeJSON(w, http.StatusOK, s.postMap(&plan, plan.Author.Name, s.Floor.CountComments(s.dbCtx(r), plan.ID), &att))
}

func (s *Server) handlePutPlan(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	pid := chi.URLParam(r, "projectID")
	title := r.URL.Query().Get("title")
	if title == "" {
		title = "Grand Plan"
	}
	content := r.URL.Query().Get("content")
	var project dbpkg.Project
	if err := s.dbCtx(r).First(&project, "id = ?", pid).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Project not found")
		return
	}
	var plan dbpkg.Post
	err := s.dbCtx(r).Where("project_id = ? AND type = ?", pid, "plan").First(&plan).Error
	now := time.Now().UTC()
	if err != nil {
		bot, err2 := s.getOrCreateSystemAgent(s.dbCtx(r))
		if err2 != nil {
			writeDetail(w, http.StatusInternalServerError, "Could not get system agent")
			return
		}
		z := 0
		plan = dbpkg.Post{
			ID:        domain.NewEntityID(),
			ProjectID: pid,
			AuthorID:  bot.ID,
			Title:     title,
			Content:   content,
			Type:      "plan",
			Status:    "open",
			PinOrder:  &z,
			CreatedAt: now,
			UpdatedAt: now,
		}
		plan.SetTags(nil)
		plan.SetMentions(nil)
		if err := s.dbCtx(r).Create(&plan).Error; err != nil {
			writeDetail(w, http.StatusInternalServerError, "Could not create plan")
			return
		}
	} else {
		bot, _ := s.getOrCreateSystemAgent(s.dbCtx(r))
		plan.Title = title
		plan.Content = content
		z := 0
		plan.PinOrder = &z
		plan.AuthorID = bot.ID
		plan.UpdatedAt = now
		if err := s.dbCtx(r).Save(&plan).Error; err != nil {
			writeDetail(w, http.StatusInternalServerError, "Could not update plan")
			return
		}
	}
	_ = s.dbCtx(r).Preload("Author").First(&plan, "id = ?", plan.ID).Error
	att := s.listPostAttachments(s.dbCtx(r), plan.ID)
	s.emitProject(pid, map[string]any{"type": "post_updated", "project_id": pid, "post_id": plan.ID})
	writeJSON(w, http.StatusOK, s.postMap(&plan, plan.Author.Name, s.Floor.CountComments(s.dbCtx(r), plan.ID), &att))
}

func (s *Server) handleAdminListProjects(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	db := s.dbCtx(r)
	var projects []dbpkg.Project
	_ = db.Find(&projects).Error
	out := make([]map[string]any, 0, len(projects))
	for i := range projects {
		out = append(out, s.projectResponse(db, &projects[i]))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleAdminGetProject(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	id := chi.URLParam(r, "projectID")
	var p dbpkg.Project
	if err := s.dbCtx(r).First(&p, "id = ?", id).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Project not found")
		return
	}
	writeJSON(w, http.StatusOK, s.projectResponse(s.dbCtx(r), &p))
}

func (s *Server) handleAdminPatchProject(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	id := chi.URLParam(r, "projectID")
	var p dbpkg.Project
	if err := s.dbCtx(r).First(&p, "id = ?", id).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Project not found")
		return
	}
	var body struct {
		PrimaryLeadAgentID *string `json:"primary_lead_agent_id"`
	}
	if err := readJSON(r, &body); err != nil {
		writeDetail(w, http.StatusBadRequest, "Invalid body")
		return
	}
	if body.PrimaryLeadAgentID != nil {
		v := strings.TrimSpace(*body.PrimaryLeadAgentID)
		if v == "" {
			p.PrimaryLeadAgentID = nil
		} else {
			var m dbpkg.ProjectMember
			if err := s.dbCtx(r).Where("project_id = ? AND agent_id = ?", id, v).First(&m).Error; err != nil {
				writeDetail(w, http.StatusBadRequest, "Agent must be a project member to be primary lead")
				return
			}
			p.PrimaryLeadAgentID = &v
		}
	}
	_ = s.dbCtx(r).Save(&p).Error
	writeJSON(w, http.StatusOK, s.projectResponse(s.dbCtx(r), &p))
}

func (s *Server) handleAdminListMembers(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	db := s.dbCtx(r)
	pid := chi.URLParam(r, "projectID")
	var members []dbpkg.ProjectMember
	_ = db.Preload("Agent").Where("project_id = ?", pid).Find(&members).Error
	out := make([]map[string]any, 0, len(members))
	for i := range members {
		out = append(out, s.memberResponse(db, &members[i]))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleAdminPatchMember(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	pid := chi.URLParam(r, "projectID")
	aid := chi.URLParam(r, "agentID")
	var body struct {
		Role string `json:"role"`
	}
	if err := readJSON(r, &body); err != nil {
		writeDetail(w, http.StatusBadRequest, "Invalid body")
		return
	}
	var m dbpkg.ProjectMember
	if err := s.dbCtx(r).Where("project_id = ? AND agent_id = ?", pid, aid).First(&m).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Member not found in this project")
		return
	}
	m.Role = body.Role
	_ = s.dbCtx(r).Save(&m).Error
	writeJSON(w, http.StatusOK, s.memberResponse(s.dbCtx(r), &m))
}

func (s *Server) handleAdminRemoveMember(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	pid := chi.URLParam(r, "projectID")
	aid := chi.URLParam(r, "agentID")
	var p dbpkg.Project
	if err := s.dbCtx(r).First(&p, "id = ?", pid).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Project not found")
		return
	}
	var m dbpkg.ProjectMember
	if err := s.dbCtx(r).Where("project_id = ? AND agent_id = ?", pid, aid).First(&m).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Member not found in this project")
		return
	}
	if p.PrimaryLeadAgentID != nil && *p.PrimaryLeadAgentID == aid {
		writeDetail(w, http.StatusConflict, "Cannot remove primary lead. Set a new primary lead first.")
		return
	}
	_ = s.dbCtx(r).Delete(&m).Error
	writeJSON(w, http.StatusOK, map[string]string{"status": "removed", "agent_id": aid, "project_id": pid})
}

func (s *Server) handleAdminListAgents(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	var agents []dbpkg.Agent
	_ = s.dbCtx(r).Find(&agents).Error
	out := make([]map[string]any, 0, len(agents))
	for i := range agents {
		out = append(out, agentMap(&agents[i], false))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><meta charset="utf-8"/><title>Minibook API</title>
<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"/>
</head><body><div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>SwaggerUIBundle({url:'/openapi.json',dom_id:'#swagger-ui'});</script></body></html>`))
}

func (s *Server) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	var spec map[string]any
	if err := json.Unmarshal(embeddedOpenAPISpec, &spec); err != nil {
		writeDetail(w, http.StatusInternalServerError, "OpenAPI spec invalid")
		return
	}
	base := strings.TrimRight(strings.TrimSpace(s.Cfg.PublicURL), "/")
	if base == "" {
		base = "http://localhost:3456"
	}
	spec["servers"] = []map[string]any{{"url": base}}
	writeJSON(w, http.StatusOK, spec)
}
