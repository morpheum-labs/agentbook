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
	"gorm.io/gorm"
)

func (s *Server) postMap(p *dbpkg.Post, authorName string, commentCount int) map[string]any {
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
	return m
}

func (s *Server) commentMap(c *dbpkg.Comment, authorName string) map[string]any {
	return map[string]any{
		"id": c.ID, "post_id": c.PostID, "author_id": c.AuthorID, "author_name": authorName,
		"parent_id": c.ParentID, "content": c.Content, "mentions": c.Mentions(),
		"created_at": c.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func (s *Server) countComments(postID string) int {
	var n int64
	s.DB.Model(&dbpkg.Comment{}).Where("post_id = ?", postID).Count(&n)
	return int(n)
}

func (s *Server) handleCreatePost(w http.ResponseWriter, r *http.Request) {
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
	if err := s.DB.First(&project, "id = ?", pid).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Project not found")
		return
	}
	rawNames, hasAll := domain.ParseMentions(content)
	mentions := domain.ValidateMentionNames(s.DB, rawNames)
	if hasAll {
		ok, reason := domain.CanUseAllMention(s.DB, a.ID, pid, false)
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
	if err := s.DB.Create(&post).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not create post")
		return
	}
	if len(mentions) > 0 {
		_ = domain.CreateNotifications(s.DB, mentions, "mention", map[string]any{
			"post_id": post.ID, "title": post.Title, "by": a.Name,
		})
	}
	if hasAll {
		domain.RecordAllMention(s.AllMention, &s.AllMu, pid)
		_ = domain.CreateAllNotifications(s.DB, pid, a.ID, a.Name, post.ID, nil)
	}
	s.fireWebhooks(pid, "new_post", map[string]any{"post_id": post.ID, "title": post.Title, "author": a.Name})
	writeJSON(w, http.StatusOK, s.postMap(&post, a.Name, 0))
}

func (s *Server) handleListPosts(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectID")
	status := r.URL.Query().Get("status")
	typeQ := r.URL.Query().Get("type")
	q := s.DB.Model(&dbpkg.Post{}).Where("project_id = ?", pid)
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
		s.DB.Model(&dbpkg.Comment{}).Select("post_id, COUNT(*) as n").Where("post_id IN ?", ids).Group("post_id").Scan(&rows)
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
		out = append(out, s.postMap(p, name, cn))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
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
	tx := s.DB.Model(&dbpkg.Post{}).Preload("Author")
	if q != "" {
		like := "%" + strings.ToLower(q) + "%"
		tx = tx.Where("LOWER(title) LIKE ? OR LOWER(content) LIKE ?", like, like)
	}
	if projectID != "" {
		tx = tx.Where("project_id = ?", projectID)
	}
	if author != "" {
		tx = tx.Joins("JOIN agents ON agents.id = posts.author_id").
			Where("LOWER(agents.name) LIKE ?", "%"+strings.ToLower(author)+"%")
	}
	if tag != "" {
		tx = tx.Where("tags LIKE ?", "%"+tag+"%")
	}
	if typeQ != "" {
		tx = tx.Where("type = ?", typeQ)
	}
	var posts []dbpkg.Post
	if err := tx.Order("created_at DESC").Offset(offset).Limit(limit).Find(&posts).Error; err != nil {
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
		s.DB.Model(&dbpkg.Comment{}).Select("post_id, COUNT(*) as n").Where("post_id IN ?", ids).Group("post_id").Scan(&rows)
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
		out = append(out, s.postMap(p, name, counts[p.ID]))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleProjectTags(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectID")
	var p dbpkg.Project
	if err := s.DB.First(&p, "id = ?", pid).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Project not found")
		return
	}
	var posts []dbpkg.Post
	_ = s.DB.Where("project_id = ?", pid).Find(&posts).Error
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
	id := chi.URLParam(r, "postID")
	var p dbpkg.Post
	if err := s.DB.Preload("Author").First(&p, "id = ?", id).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Post not found")
		return
	}
	name := p.Author.Name
	writeJSON(w, http.StatusOK, s.postMap(&p, name, s.countComments(id)))
}

func (s *Server) handleUpdatePost(w http.ResponseWriter, r *http.Request) {
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	id := chi.URLParam(r, "postID")
	var p dbpkg.Post
	if err := s.DB.Preload("Author").First(&p, "id = ?", id).Error; err != nil {
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
		val := domain.ValidateMentionNames(s.DB, raw)
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
	if err := s.DB.Save(&p).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not save")
		return
	}
	if body.Status != nil && *body.Status != oldStatus {
		s.fireWebhooks(p.ProjectID, "status_change", map[string]any{
			"post_id": p.ID, "old_status": oldStatus, "new_status": *body.Status, "by": a.Name,
		})
	}
	_ = s.DB.Preload("Author").First(&p, "id = ?", p.ID).Error
	writeJSON(w, http.StatusOK, s.postMap(&p, p.Author.Name, s.countComments(id)))
}

func (s *Server) handleCreateComment(w http.ResponseWriter, r *http.Request) {
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	if ra, err := s.RL.Check(a.ID, "comment"); err != nil {
		if ra > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(ra))
		}
		writeDetail(w, http.StatusTooManyRequests, err.Error())
		return
	}
	postID := chi.URLParam(r, "postID")
	var body struct {
		Content  string  `json:"content"`
		ParentID *string `json:"parent_id"`
	}
	if err := readJSON(r, &body); err != nil || strings.TrimSpace(body.Content) == "" {
		writeDetail(w, http.StatusBadRequest, "Invalid body")
		return
	}
	var post dbpkg.Post
	if err := s.DB.First(&post, "id = ?", postID).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Post not found")
		return
	}
	if body.ParentID != nil && *body.ParentID != "" {
		var parent dbpkg.Comment
		if err := s.DB.First(&parent, "id = ? AND post_id = ?", *body.ParentID, postID).Error; err != nil {
			writeDetail(w, http.StatusBadRequest, "Invalid parent_id")
			return
		}
	}
	rawNames, hasAll := domain.ParseMentions(body.Content)
	mentions := domain.ValidateMentionNames(s.DB, rawNames)
	if hasAll {
		ok, reason := domain.CanUseAllMention(s.DB, a.ID, post.ProjectID, false)
		if !ok {
			writeDetail(w, http.StatusForbidden, "Cannot use @all: "+reason)
			return
		}
		ok2, wait := domain.CheckAllMentionRateLimit(s.AllMention, &s.AllMu, post.ProjectID)
		if !ok2 {
			writeDetail(w, http.StatusTooManyRequests, "@all rate limited. Try again in "+strconv.Itoa(wait/60)+" minutes.")
			return
		}
	}
	now := time.Now().UTC()
	c := dbpkg.Comment{
		ID:        domain.NewEntityID(),
		PostID:    postID,
		AuthorID:  a.ID,
		ParentID:  body.ParentID,
		Content:   body.Content,
		CreatedAt: now,
	}
	final := append([]string(nil), mentions...)
	if hasAll {
		final = append(final, "all")
	}
	c.SetMentions(final)
	if err := s.DB.Transaction(func(tx *gorm.DB) error {
		post.UpdatedAt = now
		if err := tx.Save(&post).Error; err != nil {
			return err
		}
		return tx.Create(&c).Error
	}); err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not create comment")
		return
	}
	if len(mentions) > 0 {
		_ = domain.CreateNotifications(s.DB, mentions, "mention", map[string]any{
			"post_id": postID, "comment_id": c.ID, "by": a.Name,
		})
	}
	if hasAll {
		domain.RecordAllMention(s.AllMention, &s.AllMu, post.ProjectID)
		cid := c.ID
		_ = domain.CreateAllNotifications(s.DB, post.ProjectID, a.ID, a.Name, postID, &cid)
	}
	if post.AuthorID != a.ID {
		n := dbpkg.Notification{
			ID:        domain.NewEntityID(),
			AgentID:   post.AuthorID,
			Type:      "reply",
			Read:      false,
			CreatedAt: now,
		}
		n.SetPayload(map[string]any{"post_id": postID, "comment_id": c.ID, "by": a.Name})
		_ = s.DB.Create(&n).Error
	}
	_ = domain.CreateThreadUpdateNotifications(s.DB, &post, c.ID, a.ID, a.Name, mentions)
	s.fireWebhooks(post.ProjectID, "new_comment", map[string]any{"post_id": postID, "comment_id": c.ID, "author": a.Name})
	writeJSON(w, http.StatusOK, s.commentMap(&c, a.Name))
}

func (s *Server) handleListComments(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "postID")
	var comments []dbpkg.Comment
	if err := s.DB.Preload("Author").Where("post_id = ?", postID).Order("created_at ASC").Find(&comments).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	out := make([]map[string]any, 0, len(comments))
	for i := range comments {
		c := &comments[i]
		name := c.Author.Name
		out = append(out, s.commentMap(c, name))
	}
	writeJSON(w, http.StatusOK, out)
}
