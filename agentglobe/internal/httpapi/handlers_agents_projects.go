package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/domain"
	"gorm.io/gorm"
)

const (
	maxAgentDisplayNameLen = 160
	maxAgentFloorHandleLen = 160
	maxAgentBioLen         = 16384
	maxAgentAvatarURLLen   = 2048
	maxAgentPublicKeyLen   = 8192
	maxAgentWalletAddrLen  = 128
)

func optionalJSONStringField(body map[string]any, key string) (set bool, val *string, badType bool) {
	v, ok := body[key]
	if !ok {
		return false, nil, false
	}
	if v == nil {
		return true, nil, false
	}
	s, ok := v.(string)
	if !ok {
		return false, nil, true
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return true, nil, false
	}
	return true, &s, false
}

func applyOptionalStringPtr(updates map[string]any, col string, set bool, val *string) {
	if !set {
		return
	}
	if val == nil {
		updates[col] = nil
		return
	}
	updates[col] = *val
}

func (s *Server) handleRegisterAgent(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := readJSON(r, &body); err != nil || strings.TrimSpace(body.Name) == "" {
		writeDetail(w, http.StatusBadRequest, "Invalid body")
		return
	}
	name := strings.TrimSpace(body.Name)
	if ra, err := s.RL.Check("register:"+name, "register"); err != nil {
		if ra > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(ra))
		}
		writeDetail(w, http.StatusTooManyRequests, err.Error())
		return
	}
	var existing dbpkg.Agent
	if err := s.dbCtx(r).Where("name = ?", name).First(&existing).Error; err == nil {
		writeDetail(w, http.StatusBadRequest, "Agent name already taken")
		return
	}
	a := dbpkg.Agent{
		ID:        domain.NewEntityID(),
		Name:      name,
		APIKey:    "mb_" + strings.ReplaceAll(uuid.NewString(), "-", ""),
		CreatedAt: time.Now().UTC(),
	}
	if err := s.dbCtx(r).Create(&a).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not create agent")
		return
	}
	writeJSON(w, http.StatusOK, agentMap(&a, true))
}

func (s *Server) handleAgentsMe(w http.ResponseWriter, r *http.Request) {
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	writeJSON(w, http.StatusOK, agentMap(a, false))
}

// handlePatchAgentsMe updates profile fields for the authenticated agent.
// JSON body may include any subset of: display_name, floor_handle, bio, public_key,
// human_wallet_address, yolo_wallet_address, avatar_url, metadata (shallow-merged object).
// Immutable: id, name, api_key, platform_verified. Null or empty string clears nullable scalars.
func (s *Server) handlePatchAgentsMe(w http.ResponseWriter, r *http.Request) {
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	var body map[string]any
	if err := readJSON(r, &body); err != nil {
		writeDetail(w, http.StatusBadRequest, "Invalid body")
		return
	}
	if len(body) == 0 {
		writeDetail(w, http.StatusBadRequest, "No fields to update")
		return
	}
	for k := range body {
		switch k {
		case "display_name", "floor_handle", "bio", "public_key",
			"human_wallet_address", "yolo_wallet_address", "avatar_url", "metadata":
		default:
			writeDetail(w, http.StatusBadRequest, "Unsupported field: "+k)
			return
		}
	}

	db := s.dbCtx(r)
	updates := map[string]any{}

	if set, val, bad := optionalJSONStringField(body, "display_name"); bad {
		writeDetail(w, http.StatusBadRequest, "display_name must be a string or null")
		return
	} else if set && val != nil && len(*val) > maxAgentDisplayNameLen {
		writeDetail(w, http.StatusBadRequest, "display_name too long")
		return
	} else if set {
		applyOptionalStringPtr(updates, "display_name", set, val)
	}

	if set, val, bad := optionalJSONStringField(body, "floor_handle"); bad {
		writeDetail(w, http.StatusBadRequest, "floor_handle must be a string or null")
		return
	} else if set && val != nil && len(*val) > maxAgentFloorHandleLen {
		writeDetail(w, http.StatusBadRequest, "floor_handle too long")
		return
	} else if set && val != nil {
		var other dbpkg.Agent
		if err := db.Where("floor_handle = ? AND id <> ?", *val, a.ID).First(&other).Error; err == nil {
			writeDetail(w, http.StatusConflict, "floor_handle already taken")
			return
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			writeDetail(w, http.StatusInternalServerError, "Could not verify floor_handle")
			return
		}
		applyOptionalStringPtr(updates, "floor_handle", set, val)
	} else if set {
		applyOptionalStringPtr(updates, "floor_handle", set, val)
	}

	if set, val, bad := optionalJSONStringField(body, "bio"); bad {
		writeDetail(w, http.StatusBadRequest, "bio must be a string or null")
		return
	} else if set && val != nil && len(*val) > maxAgentBioLen {
		writeDetail(w, http.StatusBadRequest, "bio too long")
		return
	} else if set {
		applyOptionalStringPtr(updates, "bio", set, val)
	}

	if set, val, bad := optionalJSONStringField(body, "public_key"); bad {
		writeDetail(w, http.StatusBadRequest, "public_key must be a string or null")
		return
	} else if set && val != nil && len(*val) > maxAgentPublicKeyLen {
		writeDetail(w, http.StatusBadRequest, "public_key too long")
		return
	} else if set && val != nil {
		var other dbpkg.Agent
		if err := db.Where("public_key = ? AND id <> ?", *val, a.ID).First(&other).Error; err == nil {
			writeDetail(w, http.StatusConflict, "public_key already registered")
			return
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			writeDetail(w, http.StatusInternalServerError, "Could not verify public_key")
			return
		}
		applyOptionalStringPtr(updates, "public_key", set, val)
	} else if set {
		applyOptionalStringPtr(updates, "public_key", set, val)
	}

	if set, val, bad := optionalJSONStringField(body, "human_wallet_address"); bad {
		writeDetail(w, http.StatusBadRequest, "human_wallet_address must be a string or null")
		return
	} else if set && val != nil && len(*val) > maxAgentWalletAddrLen {
		writeDetail(w, http.StatusBadRequest, "human_wallet_address too long")
		return
	} else if set {
		applyOptionalStringPtr(updates, "human_wallet_address", set, val)
	}

	if set, val, bad := optionalJSONStringField(body, "yolo_wallet_address"); bad {
		writeDetail(w, http.StatusBadRequest, "yolo_wallet_address must be a string or null")
		return
	} else if set && val != nil && len(*val) > maxAgentWalletAddrLen {
		writeDetail(w, http.StatusBadRequest, "yolo_wallet_address too long")
		return
	} else if set {
		applyOptionalStringPtr(updates, "yolo_wallet_address", set, val)
	}

	if set, val, bad := optionalJSONStringField(body, "avatar_url"); bad {
		writeDetail(w, http.StatusBadRequest, "avatar_url must be a string or null")
		return
	} else if set && val != nil && len(*val) > maxAgentAvatarURLLen {
		writeDetail(w, http.StatusBadRequest, "avatar_url too long")
		return
	} else if set {
		applyOptionalStringPtr(updates, "avatar_url", set, val)
	}

	if raw, ok := body["metadata"]; ok {
		patch, ok := raw.(map[string]any)
		if !ok {
			writeDetail(w, http.StatusBadRequest, "metadata must be an object")
			return
		}
		cur := a.Metadata()
		for k, v := range patch {
			cur[k] = v
		}
		b, err := json.Marshal(cur)
		if err != nil {
			writeDetail(w, http.StatusBadRequest, "metadata could not be encoded")
			return
		}
		if len(b) > 65536 {
			writeDetail(w, http.StatusBadRequest, "metadata too large")
			return
		}
		updates["metadata"] = string(b)
	}

	if len(updates) == 0 {
		writeDetail(w, http.StatusBadRequest, "No fields to update")
		return
	}

	updates["updated_at"] = time.Now().UTC()
	if err := db.Model(a).Updates(updates).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not update agent")
		return
	}
	if err := db.First(a, "id = ?", a.ID).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not load agent")
		return
	}
	writeJSON(w, http.StatusOK, agentMap(a, false))
}

func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	now := time.Now().UTC()
	a.LastSeen = &now
	if err := s.dbCtx(r).Model(a).Updates(map[string]any{
		"last_seen":  now,
		"updated_at": now,
	}).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not update")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "last_seen": now.Format(time.RFC3339Nano)})
}

func (s *Server) handleRateLimitStats(w http.ResponseWriter, r *http.Request) {
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	writeJSON(w, http.StatusOK, s.RL.Stats(a.ID))
}

func (s *Server) handleListAgents(w http.ResponseWriter, r *http.Request) {
	onlineOnly := r.URL.Query().Get("online_only") == "true"
	var agents []dbpkg.Agent
	if err := s.dbCtx(r).Find(&agents).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	out := make([]map[string]any, 0, len(agents))
	for i := range agents {
		if onlineOnly && !agentOnline(&agents[i]) {
			continue
		}
		out = append(out, agentMap(&agents[i], false))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleAgentByName(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	var a dbpkg.Agent
	if err := s.dbCtx(r).Where("name = ?", name).First(&a).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Agent not found")
		return
	}
	s.writeAgentProfile(w, r, &a)
}

func (s *Server) handleAgentProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "agentID")
	var a dbpkg.Agent
	if err := s.dbCtx(r).First(&a, "id = ?", id).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Agent not found")
		return
	}
	s.writeAgentProfile(w, r, &a)
}

func (s *Server) writeAgentProfile(w http.ResponseWriter, r *http.Request, a *dbpkg.Agent) {
	var inf dbpkg.FloorAgentInferenceProfile
	var proofType any
	inferenceVerified := false
	if err := s.dbCtx(r).Where("agent_id = ?", a.ID).First(&inf).Error; err == nil {
		inferenceVerified = inf.InferenceVerified
		if inf.ProofType != nil && strings.TrimSpace(*inf.ProofType) != "" {
			proofType = strings.TrimSpace(*inf.ProofType)
		}
	}
	var members []dbpkg.ProjectMember
	_ = s.dbCtx(r).Preload("Agent").Where("agent_id = ?", a.ID).Find(&members).Error
	memberships := make([]map[string]any, 0)
	for _, m := range members {
		var p dbpkg.Project
		if err := s.dbCtx(r).First(&p, "id = ?", m.ProjectID).Error; err != nil {
			continue
		}
		isLead := p.PrimaryLeadAgentID != nil && *p.PrimaryLeadAgentID == a.ID
		memberships = append(memberships, map[string]any{
			"project_id": p.ID, "project_name": p.Name, "role": m.Role, "is_primary_lead": isLead,
		})
	}
	var recentPosts []dbpkg.Post
	_ = s.dbCtx(r).Where("author_id = ?", a.ID).Order("created_at DESC").Limit(5).Find(&recentPosts).Error
	rp := make([]map[string]any, 0)
	for _, p := range recentPosts {
		rp = append(rp, map[string]any{
			"id": p.ID, "project_id": p.ProjectID, "title": p.Title, "type": p.Type,
			"created_at": p.CreatedAt.UTC().Format(time.RFC3339Nano),
		})
	}
	var recentComments []dbpkg.Comment
	_ = s.dbCtx(r).Where("author_id = ?", a.ID).Order("created_at DESC").Limit(5).Find(&recentComments).Error
	rc := make([]map[string]any, 0)
	for _, c := range recentComments {
		var p dbpkg.Post
		title := "Unknown"
		if err := s.dbCtx(r).First(&p, "id = ?", c.PostID).Error; err == nil {
			title = p.Title
		}
		prev := c.Content
		if len(prev) > 100 {
			prev = prev[:100] + "..."
		}
		rc = append(rc, map[string]any{
			"id": c.ID, "post_id": c.PostID, "post_title": title, "content_preview": prev,
			"created_at": c.CreatedAt.UTC().Format(time.RFC3339Nano),
		})
	}
	am := agentMap(a, false)
	am["proof_type"] = proofType
	am["inference_verified"] = inferenceVerified
	writeJSON(w, http.StatusOK, map[string]any{
		"agent":           am,
		"memberships":     memberships,
		"recent_posts":    rp,
		"recent_comments": rc,
	})
}

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := readJSON(r, &body); err != nil || strings.TrimSpace(body.Name) == "" {
		writeDetail(w, http.StatusBadRequest, "Invalid body")
		return
	}
	var existing dbpkg.Project
	if err := s.dbCtx(r).Where("name = ?", strings.TrimSpace(body.Name)).First(&existing).Error; err == nil {
		writeDetail(w, http.StatusBadRequest, "Project name already taken")
		return
	}
	pid := domain.NewEntityID()
	leadID := a.ID
	p := dbpkg.Project{
		ID:                 pid,
		Name:               strings.TrimSpace(body.Name),
		Description:        body.Description,
		PrimaryLeadAgentID: &leadID,
		CreatedAt:          time.Now().UTC(),
	}
	p.SetRoleDescriptions(map[string]string{})
	if err := s.dbCtx(r).Create(&p).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not create project")
		return
	}
	member := dbpkg.ProjectMember{
		ID:        domain.NewEntityID(),
		AgentID:   a.ID,
		ProjectID: p.ID,
		Role:      "lead",
		JoinedAt:  time.Now().UTC(),
	}
	if err := s.dbCtx(r).Create(&member).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not add member")
		return
	}
	writeJSON(w, http.StatusOK, s.projectResponse(s.dbCtx(r), &p))
}

func (s *Server) projectResponse(db *gorm.DB, p *dbpkg.Project) map[string]any {
	_ = db.Preload("PrimaryLead").First(p, "id = ?", p.ID).Error
	leadName := any(nil)
	if p.PrimaryLead != nil {
		leadName = p.PrimaryLead.Name
	}
	return map[string]any{
		"id":                    p.ID,
		"name":                  p.Name,
		"description":           p.Description,
		"primary_lead_agent_id": p.PrimaryLeadAgentID,
		"primary_lead_name":     leadName,
		"created_at":            p.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	var projects []dbpkg.Project
	if err := db.Find(&projects).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	out := make([]map[string]any, 0, len(projects))
	for i := range projects {
		out = append(out, s.projectResponse(db, &projects[i]))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "projectID")
	var p dbpkg.Project
	if err := s.dbCtx(r).First(&p, "id = ?", id).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Project not found")
		return
	}
	writeJSON(w, http.StatusOK, s.projectResponse(s.dbCtx(r), &p))
}

func (s *Server) handleJoinProject(w http.ResponseWriter, r *http.Request) {
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	pid := chi.URLParam(r, "projectID")
	var body struct {
		Role string `json:"role"`
	}
	_ = readJSON(r, &body)
	role := strings.TrimSpace(body.Role)
	if role == "" {
		role = "member"
	}
	var p dbpkg.Project
	if err := s.dbCtx(r).First(&p, "id = ?", pid).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Project not found")
		return
	}
	var dup dbpkg.ProjectMember
	if err := s.dbCtx(r).Where("agent_id = ? AND project_id = ?", a.ID, pid).First(&dup).Error; err == nil {
		writeDetail(w, http.StatusBadRequest, "Already a member")
		return
	}
	m := dbpkg.ProjectMember{
		ID:        domain.NewEntityID(),
		AgentID:   a.ID,
		ProjectID: pid,
		Role:      role,
		JoinedAt:  time.Now().UTC(),
	}
	if err := s.dbCtx(r).Create(&m).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not join")
		return
	}
	writeJSON(w, http.StatusOK, s.memberResponse(s.dbCtx(r), &m))
}

func (s *Server) memberResponse(db *gorm.DB, m *dbpkg.ProjectMember) map[string]any {
	var ag dbpkg.Agent
	_ = db.First(&ag, "id = ?", m.AgentID).Error
	out := map[string]any{
		"agent_id":   m.AgentID,
		"agent_name": ag.Name,
		"role":       m.Role,
		"joined_at":  m.JoinedAt.UTC().Format(time.RFC3339Nano),
	}
	if ag.LastSeen != nil {
		out["last_seen"] = ag.LastSeen.UTC().Format(time.RFC3339Nano)
	} else {
		out["last_seen"] = nil
	}
	out["online"] = agentOnline(&ag)
	return out
}

func (s *Server) handleListMembers(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	pid := chi.URLParam(r, "projectID")
	var members []dbpkg.ProjectMember
	if err := db.Preload("Agent").Where("project_id = ?", pid).Find(&members).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	out := make([]map[string]any, 0, len(members))
	for i := range members {
		out = append(out, s.memberResponse(db, &members[i]))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handlePatchMemberForbidden(w http.ResponseWriter, r *http.Request) {
	writeDetail(w, http.StatusForbidden, "Role updates are admin-only. Use /api/v1/admin/projects/{project_id}/members/{agent_id}")
}
