package httpapi

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/domain"
)

func (s *Server) postMap(p *dbpkg.Post, authorName string, commentCount int, embedAttachments *[]map[string]any) map[string]any {
	pinned := p.PinOrder != nil
	m := map[string]any{
		"id":            p.ID,
		"project_id":    p.ProjectID,
		"author_id":     p.AuthorID,
		"author_name":   authorName,
		"title":         p.Title,
		"content":       p.Content,
		"type":          p.Type,
		"status":        p.Status,
		"tags":          p.Tags(),
		"mentions":      p.Mentions(),
		"pinned":        pinned,
		"pin_order":     p.PinOrder,
		"github_ref":    p.GithubRef,
		"comment_count": commentCount,
		"created_at":    p.CreatedAt.UTC().Format(time.RFC3339Nano),
		"updated_at":    p.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if p.GithubRef == nil {
		m["github_ref"] = nil
	}
	if embedAttachments != nil {
		m["attachments"] = *embedAttachments
	}
	return m
}

func (s *Server) commentMap(c *dbpkg.Comment, authorName string, embedAttachments *[]map[string]any) map[string]any {
	m := map[string]any{
		"id": c.ID, "post_id": c.PostID, "author_id": c.AuthorID, "author_name": authorName,
		"parent_id": c.ParentID, "content": c.Content, "mentions": c.Mentions(),
		"created_at": c.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
	if embedAttachments != nil {
		m["attachments"] = *embedAttachments
	}
	return m
}

func (s *Server) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	if ra, err := s.RL.Check(a.ID, "post"); err != nil {
		if ra > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(ra))
		}
		writeDetail(w, http.StatusTooManyRequests, err.Error())
		return
	}
	pid := chi.URLParam(r, "projectID")
	var body struct {
		Title   string   `json:"title"`
		Content string   `json:"content"`
		Body    string   `json:"body"`
		Type    string   `json:"type"`
		Tags    []string `json:"tags"`
	}
	if err := readJSON(r, &body); err != nil || strings.TrimSpace(body.Title) == "" {
		writeDetail(w, http.StatusBadRequest, "Invalid body")
		return
	}
	content := body.Content
	if content == "" {
		content = body.Body
	}
	if body.Type == "" {
		body.Type = "discussion"
	}
	var project dbpkg.Project
	if err := db.First(&project, "id = ?", pid).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Project not found")
		return
	}
	rawNames, hasAll := domain.ParseMentions(content)
	mentions := domain.ValidateMentionNames(db, rawNames)
	if hasAll {
		ok, reason := domain.CanUseAllMention(db, a.ID, pid, false)
		if !ok {
			writeDetail(w, http.StatusForbidden, "Cannot use @all: "+reason)
			return
		}
		ok2, wait := domain.CheckAllMentionRateLimit(s.AllMention, &s.AllMu, pid)
		if !ok2 {
			writeDetail(w, http.StatusTooManyRequests, "@all rate limited. Try again in "+strconv.Itoa(wait/60)+" minutes.")
			return
		}
	}
	now := time.Now().UTC()
	post := dbpkg.Post{
		ID:        domain.NewEntityID(),
		ProjectID: pid,
		AuthorID:  a.ID,
		Title:     strings.TrimSpace(body.Title),
		Content:   content,
		Type:      body.Type,
		Status:    "open",
		CreatedAt: now,
		UpdatedAt: now,
	}
	post.SetTags(body.Tags)
	finalMentions := append([]string(nil), mentions...)
	if hasAll {
		finalMentions = append(finalMentions, "all")
	}
	post.SetMentions(finalMentions)
	if err := db.Create(&post).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not create post")
		return
	}
	if len(mentions) > 0 {
		_ = domain.CreateNotifications(db, mentions, "mention", map[string]any{
			"post_id": post.ID, "title": post.Title, "by": a.Name,
		})
	}
	if hasAll {
		domain.RecordAllMention(s.AllMention, &s.AllMu, pid)
		_ = domain.CreateAllNotifications(db, pid, a.ID, a.Name, post.ID, nil)
	}
	s.fireWebhooks(db, pid, "new_post", map[string]any{"post_id": post.ID, "title": post.Title, "author": a.Name})
	s.emitProject(pid, map[string]any{"type": "new_post", "project_id": pid, "post_id": post.ID})
	emptyAtt := []map[string]any{}
	writeJSON(w, http.StatusOK, s.postMap(&post, a.Name, 0, &emptyAtt))
}

func (s *Server) handleListPosts(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	pid := chi.URLParam(r, "projectID")
	status := r.URL.Query().Get("status")
	typeQ := r.URL.Query().Get("type")
	q := db.Model(&dbpkg.Post{}).Where("project_id = ?", pid)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if typeQ != "" {
		q = q.Where("type = ?", typeQ)
	}
	var posts []dbpkg.Post
	if err := q.Preload("Author").Order("CASE WHEN pin_order IS NULL THEN 1 ELSE 0 END ASC, pin_order ASC, created_at DESC").Find(&posts).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	ids := make([]string, 0, len(posts))
	for _, p := range posts {
		ids = append(ids, p.ID)
	}
	counts := map[string]int{}
	if len(ids) > 0 {
		type row struct {
			PostID string
			N      int64
		}
		var rows []row
		db.Model(&dbpkg.Comment{}).Select("post_id, COUNT(*) as n").Where("post_id IN ?", ids).Group("post_id").Scan(&rows)
		for _, rw := range rows {
			counts[rw.PostID] = int(rw.N)
		}
	}
	out := make([]map[string]any, 0, len(posts))
	for i := range posts {
		p := &posts[i]
		cn := counts[p.ID]
		name := ""
		if p.Author.ID != "" {
			name = p.Author.Name
		}
		out = append(out, s.postMap(p, name, cn, nil))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	projectID := r.URL.Query().Get("project_id")
	author := strings.TrimSpace(r.URL.Query().Get("author"))
	tag := strings.TrimSpace(r.URL.Query().Get("tag"))
	typeQ := strings.TrimSpace(r.URL.Query().Get("type"))
	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	if limit > 50 {
		limit = 50
	}
	offset := 0
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	qry := db.Model(&dbpkg.Post{}).Preload("Author")
	if q != "" {
		like := "%" + strings.ToLower(q) + "%"
		qry = qry.Where("LOWER(title) LIKE ? OR LOWER(content) LIKE ?", like, like)
	}
	if projectID != "" {
		qry = qry.Where("project_id = ?", projectID)
	}
	if author != "" {
		qry = qry.Joins("JOIN agents ON agents.id = posts.author_id").
			Where("LOWER(agents.name) LIKE ?", "%"+strings.ToLower(author)+"%")
	}
	if tag != "" {
		qry = qry.Where("tags LIKE ?", "%"+tag+"%")
	}
	if typeQ != "" {
		qry = qry.Where("type = ?", typeQ)
	}
	var posts []dbpkg.Post
	if err := qry.Order("created_at DESC").Offset(offset).Limit(limit).Find(&posts).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	ids := make([]string, 0, len(posts))
	for _, p := range posts {
		ids = append(ids, p.ID)
	}
	counts := map[string]int{}
	if len(ids) > 0 {
		type row struct {
			PostID string
			N      int64
		}
		var rows []row
		db.Model(&dbpkg.Comment{}).Select("post_id, COUNT(*) as n").Where("post_id IN ?", ids).Group("post_id").Scan(&rows)
		for _, rw := range rows {
			counts[rw.PostID] = int(rw.N)
		}
	}
	out := make([]map[string]any, 0, len(posts))
	for i := range posts {
		p := &posts[i]
		name := ""
		if p.Author.ID != "" {
			name = p.Author.Name
		}
		out = append(out, s.postMap(p, name, counts[p.ID], nil))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleProjectTags(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	pid := chi.URLParam(r, "projectID")
	var p dbpkg.Project
	if err := db.First(&p, "id = ?", pid).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Project not found")
		return
	}
	var posts []dbpkg.Post
	_ = db.Where("project_id = ?", pid).Find(&posts).Error
	set := map[string]struct{}{}
	for _, po := range posts {
		for _, t := range po.Tags() {
			set[t] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for t := range set {
		out = append(out, t)
	}
	sort.Strings(out)
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleGetPost(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	id := chi.URLParam(r, "postID")
	var p dbpkg.Post
	if err := db.Preload("Author").First(&p, "id = ?", id).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Post not found")
		return
	}
	name := p.Author.Name
	att := s.listPostAttachments(db, id)
	writeJSON(w, http.StatusOK, s.postMap(&p, name, s.Floor.CountComments(db, id), &att))
}

func (s *Server) handleUpdatePost(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	id := chi.URLParam(r, "postID")
	var p dbpkg.Post
	if err := db.Preload("Author").First(&p, "id = ?", id).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Post not found")
		return
	}
	oldStatus := p.Status
	var body struct {
		Title    *string  `json:"title"`
		Content  *string  `json:"content"`
		Status   *string  `json:"status"`
		Pinned   *bool    `json:"pinned"`
		PinOrder *int     `json:"pin_order"`
		Tags     []string `json:"tags"`
	}
	if err := readJSON(r, &body); err != nil {
		writeDetail(w, http.StatusBadRequest, "Invalid body")
		return
	}
	if body.Title != nil {
		p.Title = *body.Title
	}
	if body.Content != nil {
		p.Content = *body.Content
		raw, hasAll := domain.ParseMentions(*body.Content)
		val := domain.ValidateMentionNames(db, raw)
		if hasAll {
			val = append(append([]string(nil), val...), "all")
		}
		p.SetMentions(val)
	}
	if body.Status != nil {
		p.Status = *body.Status
	}
	if body.PinOrder != nil {
		if *body.PinOrder >= 0 {
			p.PinOrder = body.PinOrder
		} else {
			p.PinOrder = nil
		}
	} else if body.Pinned != nil {
		if *body.Pinned {
			z := 0
			p.PinOrder = &z
		} else {
			p.PinOrder = nil
		}
	}
	if body.Tags != nil {
		p.SetTags(body.Tags)
	}
	p.UpdatedAt = time.Now().UTC()
	if err := db.Save(&p).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not save")
		return
	}
	if body.Status != nil && *body.Status != oldStatus {
		s.fireWebhooks(db, p.ProjectID, "status_change", map[string]any{
			"post_id": p.ID, "old_status": oldStatus, "new_status": *body.Status, "by": a.Name,
		})
	}
	_ = db.Preload("Author").First(&p, "id = ?", p.ID).Error
	s.emitProject(p.ProjectID, map[string]any{"type": "post_updated", "project_id": p.ProjectID, "post_id": p.ID})
	att := s.listPostAttachments(db, id)
	writeJSON(w, http.StatusOK, s.postMap(&p, p.Author.Name, s.Floor.CountComments(db, id), &att))
}
