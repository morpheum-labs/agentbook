package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/domain"
	"gorm.io/gorm"
)

func (s *Server) handleCreateComment(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
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
	if err := db.First(&post, "id = ?", postID).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Post not found")
		return
	}
	if body.ParentID != nil && *body.ParentID != "" {
		var parent dbpkg.Comment
		if err := db.First(&parent, "id = ? AND post_id = ?", *body.ParentID, postID).Error; err != nil {
			writeDetail(w, http.StatusBadRequest, "Invalid parent_id")
			return
		}
	}
	rawNames, hasAll := domain.ParseMentions(body.Content)
	mentions := domain.ValidateMentionNames(db, rawNames)
	if hasAll {
		ok, reason := domain.CanUseAllMention(db, a.ID, post.ProjectID, false)
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
	if err := db.Transaction(func(tx *gorm.DB) error {
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
		_ = domain.CreateNotifications(db, mentions, "mention", map[string]any{
			"post_id": postID, "comment_id": c.ID, "by": a.Name,
		})
	}
	if hasAll {
		domain.RecordAllMention(s.AllMention, &s.AllMu, post.ProjectID)
		cid := c.ID
		_ = domain.CreateAllNotifications(db, post.ProjectID, a.ID, a.Name, postID, &cid)
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
		_ = db.Create(&n).Error
	}
	_ = domain.CreateThreadUpdateNotifications(db, &post, c.ID, a.ID, a.Name, mentions)
	s.fireWebhooks(db, post.ProjectID, "new_comment", map[string]any{"post_id": postID, "comment_id": c.ID, "author": a.Name})
	s.emitProject(post.ProjectID, map[string]any{"type": "new_comment", "project_id": post.ProjectID, "post_id": postID, "comment_id": c.ID})
	emptyAtt := []map[string]any{}
	writeJSON(w, http.StatusOK, s.commentMap(&c, a.Name, &emptyAtt))
}

func (s *Server) handleListComments(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	postID := chi.URLParam(r, "postID")
	var comments []dbpkg.Comment
	if err := db.Preload("Author").Where("post_id = ?", postID).Order("created_at ASC").Find(&comments).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	ids := make([]string, 0, len(comments))
	for i := range comments {
		ids = append(ids, comments[i].ID)
	}
	byAtt := s.attachmentsByCommentIDs(db, ids)
	out := make([]map[string]any, 0, len(comments))
	for i := range comments {
		c := &comments[i]
		name := c.Author.Name
		list := byAtt[c.ID]
		if list == nil {
			list = []map[string]any{}
		}
		out = append(out, s.commentMap(c, name, &list))
	}
	writeJSON(w, http.StatusOK, out)
}
