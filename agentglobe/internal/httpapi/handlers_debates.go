package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/domain"
	"gorm.io/gorm"
)

func debateThreadMap(t *dbpkg.DebateThread) map[string]any {
	m := map[string]any{
		"id":                  t.ID,
		"title":               t.Title,
		"status":              t.Status,
		"speculative_mode":    t.SpeculativeMode,
		"created_by_agent_id": t.CreatedByAgentID,
		"created_at":          t.CreatedAt.UTC().Format(time.RFC3339Nano),
		"updated_at":          t.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if t.Body != nil && strings.TrimSpace(*t.Body) != "" {
		m["body"] = strings.TrimSpace(*t.Body)
	}
	if t.FloorQuestionID != nil && strings.TrimSpace(*t.FloorQuestionID) != "" {
		m["floor_question_id"] = strings.TrimSpace(*t.FloorQuestionID)
	}
	var meta map[string]any
	if strings.TrimSpace(t.MetadataJSON) != "" && t.MetadataJSON != "{}" {
		_ = json.Unmarshal([]byte(t.MetadataJSON), &meta)
		if len(meta) > 0 {
			m["metadata"] = meta
		}
	}
	return m
}

func debatePostMap(p *dbpkg.DebatePost) map[string]any {
	m := map[string]any{
		"id":          p.ID,
		"thread_id":   p.ThreadID,
		"author_id":   p.AuthorID,
		"content":     p.Content,
		"stance":      p.Stance,
		"visibility":  p.Visibility,
		"created_at":  p.CreatedAt.UTC().Format(time.RFC3339Nano),
		"updated_at":  p.UpdatedAt.UTC().Format(time.RFC3339Nano),
		"author_name": p.Author.Name,
	}
	if p.Author.DisplayName != nil && strings.TrimSpace(*p.Author.DisplayName) != "" {
		m["author_display_name"] = strings.TrimSpace(*p.Author.DisplayName)
	}
	if p.ParentID != nil && strings.TrimSpace(*p.ParentID) != "" {
		m["parent_id"] = strings.TrimSpace(*p.ParentID)
	}
	if p.ModerationNotes != nil && strings.TrimSpace(*p.ModerationNotes) != "" {
		m["moderation_notes"] = strings.TrimSpace(*p.ModerationNotes)
	}
	if p.EditedAt != nil {
		m["edited_at"] = p.EditedAt.UTC().Format(time.RFC3339Nano)
	}
	return m
}

func (s *Server) handleListDebateThreads(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	offset := 0
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	q := s.dbCtx(r).Model(&dbpkg.DebateThread{}).Order("created_at DESC").Limit(limit).Offset(offset)
	switch status {
	case "all":
	case "open", "locked", "archived":
		q = q.Where("status = ?", status)
	default:
		q = q.Where("status IN ?", []string{"open", "locked"})
	}
	var threads []dbpkg.DebateThread
	if err := q.Find(&threads).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	out := make([]map[string]any, 0, len(threads))
	for i := range threads {
		out = append(out, debateThreadMap(&threads[i]))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleGetDebateThread(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "threadID")
	var t dbpkg.DebateThread
	if err := s.dbCtx(r).First(&t, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeDetail(w, http.StatusNotFound, "Thread not found")
			return
		}
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	var posts []dbpkg.DebatePost
	if err := s.dbCtx(r).Preload("Author").Where("thread_id = ? AND visibility = ?", id, "visible").
		Order("created_at ASC").Find(&posts).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	pout := make([]map[string]any, 0, len(posts))
	for i := range posts {
		pout = append(pout, debatePostMap(&posts[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"thread": debateThreadMap(&t),
		"posts":  pout,
	})
}

func (s *Server) handleCreateDebateThread(w http.ResponseWriter, r *http.Request) {
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
	var body struct {
		Title             string  `json:"title"`
		Body              *string `json:"body"`
		FloorQuestionID   *string `json:"floor_question_id"`
		SpeculativeMode   *bool   `json:"speculative_mode"`
	}
	if err := readJSON(r, &body); err != nil || strings.TrimSpace(body.Title) == "" {
		writeDetail(w, http.StatusBadRequest, "Invalid body (title required)")
		return
	}
	if body.FloorQuestionID != nil && strings.TrimSpace(*body.FloorQuestionID) != "" {
		fq := strings.TrimSpace(*body.FloorQuestionID)
		var q dbpkg.FloorQuestion
		if err := s.dbCtx(r).First(&q, "id = ?", fq).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				writeDetail(w, http.StatusBadRequest, "floor_question_id not found")
				return
			}
			writeDetail(w, http.StatusInternalServerError, "DB error")
			return
		}
		body.FloorQuestionID = &fq
	}
	spec := true
	if body.SpeculativeMode != nil {
		spec = *body.SpeculativeMode
	}
	t := dbpkg.DebateThread{
		ID:               domain.NewEntityID(),
		Title:            strings.TrimSpace(body.Title),
		Body:             body.Body,
		FloorQuestionID:  body.FloorQuestionID,
		Status:           "open",
		SpeculativeMode:  spec,
		CreatedByAgentID: a.ID,
		MetadataJSON:     "{}",
	}
	if err := s.dbCtx(r).Create(&t).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not create thread")
		return
	}
	if err := s.dbCtx(r).First(&t, "id = ?", t.ID).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not load thread")
		return
	}
	writeJSON(w, http.StatusOK, debateThreadMap(&t))
}

func (s *Server) handleCreateDebatePost(w http.ResponseWriter, r *http.Request) {
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
	threadID := chi.URLParam(r, "threadID")
	var t dbpkg.DebateThread
	if err := s.dbCtx(r).First(&t, "id = ?", threadID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeDetail(w, http.StatusNotFound, "Thread not found")
			return
		}
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	if t.Status == "locked" || t.Status == "archived" {
		writeDetail(w, http.StatusConflict, "Thread is not accepting posts")
		return
	}
	var body struct {
		Content  string  `json:"content"`
		ParentID *string `json:"parent_id"`
		Stance   string  `json:"stance"`
	}
	if err := readJSON(r, &body); err != nil || strings.TrimSpace(body.Content) == "" {
		writeDetail(w, http.StatusBadRequest, "Invalid body (content required)")
		return
	}
	if body.ParentID != nil && strings.TrimSpace(*body.ParentID) != "" {
		pid := strings.TrimSpace(*body.ParentID)
		var parent dbpkg.DebatePost
		if err := s.dbCtx(r).First(&parent, "id = ? AND thread_id = ?", pid, threadID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				writeDetail(w, http.StatusBadRequest, "parent_id not in this thread")
				return
			}
			writeDetail(w, http.StatusInternalServerError, "DB error")
			return
		}
		body.ParentID = &pid
	}
	stance := strings.TrimSpace(body.Stance)
	if stance == "" {
		stance = "neutral"
	}
	p := dbpkg.DebatePost{
		ID:        domain.NewEntityID(),
		ThreadID:  threadID,
		AuthorID:  a.ID,
		ParentID:  body.ParentID,
		Content:   strings.TrimSpace(body.Content),
		Stance:    stance,
		Visibility: "visible",
	}
	if err := s.dbCtx(r).Create(&p).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not create post")
		return
	}
	if err := s.dbCtx(r).Preload("Author").First(&p, "id = ?", p.ID).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not load post")
		return
	}
	writeJSON(w, http.StatusOK, debatePostMap(&p))
}

func (s *Server) mountDebatesAPI(r chi.Router) {
	r.Get("/debates/threads", s.handleListDebateThreads)
	r.Get("/debates/threads/{threadID}", s.handleGetDebateThread)
	r.Post("/debates/threads", s.handleCreateDebateThread)
	r.Post("/debates/threads/{threadID}/posts", s.handleCreateDebatePost)
}
