package httpapi

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/domain"
)

func (s *Server) attachmentDir() string {
	return strings.TrimSpace(s.Cfg.AttachmentsDir)
}

func (s *Server) attachmentMeta(a *dbpkg.Attachment) map[string]any {
	m := map[string]any{
		"id":             a.ID,
		"project_id":    a.ProjectID,
		"filename":       a.Filename,
		"content_type":   a.ContentType,
		"size":           a.Size,
		"author_id":      a.AuthorID,
		"download_path":  "/api/v1/attachments/" + a.ID,
		"created_at":     a.CreatedAt.UTC().Format(time.RFC3339Nano),
		"post_id":        nil,
		"comment_id":     nil,
	}
	if a.PostID != nil {
		m["post_id"] = *a.PostID
	}
	if a.CommentID != nil {
		m["comment_id"] = *a.CommentID
	}
	return m
}

func (s *Server) listPostAttachments(postID string) []map[string]any {
	var rows []dbpkg.Attachment
	_ = s.DB.Where("post_id = ? AND comment_id IS NULL", postID).Order("created_at ASC").Find(&rows).Error
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, s.attachmentMeta(&rows[i]))
	}
	return out
}

func (s *Server) attachmentsByCommentIDs(ids []string) map[string][]map[string]any {
	if len(ids) == 0 {
		return map[string][]map[string]any{}
	}
	var rows []dbpkg.Attachment
	_ = s.DB.Where("comment_id IN ?", ids).Order("created_at ASC").Find(&rows).Error
	by := make(map[string][]map[string]any)
	for i := range rows {
		if rows[i].CommentID == nil {
			continue
		}
		cid := *rows[i].CommentID
		by[cid] = append(by[cid], s.attachmentMeta(&rows[i]))
	}
	return by
}

func (s *Server) handleListPostAttachments(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "postID")
	if err := s.DB.First(&dbpkg.Post{}, "id = ?", postID).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Post not found")
		return
	}
	writeJSON(w, http.StatusOK, s.listPostAttachments(postID))
}

func (s *Server) handleListCommentAttachments(w http.ResponseWriter, r *http.Request) {
	commentID := chi.URLParam(r, "commentID")
	var c dbpkg.Comment
	if err := s.DB.First(&c, "id = ?", commentID).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Comment not found")
		return
	}
	var rows []dbpkg.Attachment
	_ = s.DB.Where("comment_id = ?", commentID).Order("created_at ASC").Find(&rows).Error
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, s.attachmentMeta(&rows[i]))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleUploadPostAttachment(w http.ResponseWriter, r *http.Request) {
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	if ra, err := s.RL.Check(a.ID, "attachment"); err != nil {
		if ra > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(ra))
		}
		writeDetail(w, http.StatusTooManyRequests, err.Error())
		return
	}
	postID := chi.URLParam(r, "postID")
	var post dbpkg.Post
	if err := s.DB.First(&post, "id = ?", postID).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Post not found")
		return
	}
	s.saveUploadedAttachment(w, r, a, post.ProjectID, &postID, nil)
}

func (s *Server) handleUploadCommentAttachment(w http.ResponseWriter, r *http.Request) {
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	if ra, err := s.RL.Check(a.ID, "attachment"); err != nil {
		if ra > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(ra))
		}
		writeDetail(w, http.StatusTooManyRequests, err.Error())
		return
	}
	commentID := chi.URLParam(r, "commentID")
	var c dbpkg.Comment
	if err := s.DB.First(&c, "id = ?", commentID).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Comment not found")
		return
	}
	var post dbpkg.Post
	if err := s.DB.First(&post, "id = ?", c.PostID).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Post not found")
		return
	}
	pid := c.PostID
	s.saveUploadedAttachment(w, r, a, post.ProjectID, &pid, &commentID)
}

func (s *Server) saveUploadedAttachment(w http.ResponseWriter, r *http.Request, a *dbpkg.Agent, projectID string, postID, commentID *string) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeDetail(w, http.StatusBadRequest, "Invalid multipart form")
		return
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		writeDetail(w, http.StatusBadRequest, "Missing file field")
		return
	}
	defer file.Close()
	lim := s.Cfg.MaxAttachmentBytes
	if lim <= 0 {
		lim = 10 * 1024 * 1024
	}
	sniff := make([]byte, 512)
	n, _ := file.Read(sniff)
	head := sniff[:n]
	contentType := http.DetectContentType(head)
	var body io.Reader = file
	if n > 0 {
		body = io.MultiReader(bytes.NewReader(head), file)
	}
	pr := io.LimitReader(body, lim+1)
	id := domain.NewEntityID()
	dir := s.attachmentDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not create storage directory")
		return
	}
	diskPath := filepath.Join(dir, id)
	out, err := os.Create(diskPath)
	if err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not store file")
		return
	}
	written, err := io.Copy(out, pr)
	_ = out.Close()
	if err != nil {
		_ = os.Remove(diskPath)
		writeDetail(w, http.StatusInternalServerError, "Could not store file")
		return
	}
	if written > lim {
		_ = os.Remove(diskPath)
		writeDetail(w, http.StatusRequestEntityTooLarge, "File exceeds max_attachment_bytes")
		return
	}
	filename := filepath.Base(hdr.Filename)
	if filename == "" || filename == "." {
		filename = "upload"
	}
	if ext := strings.ToLower(filepath.Ext(filename)); ext != "" &&
		(contentType == "application/octet-stream" || strings.HasPrefix(contentType, "text/plain")) {
		if mt := mime.TypeByExtension(ext); mt != "" {
			contentType = mt
		}
	}
	now := time.Now().UTC()
	row := dbpkg.Attachment{
		ID:          id,
		ProjectID:   projectID,
		PostID:      postID,
		CommentID:   commentID,
		AuthorID:    a.ID,
		Filename:    filename,
		ContentType: contentType,
		Size:        written,
		CreatedAt:   now,
	}
	if err := s.DB.Create(&row).Error; err != nil {
		_ = os.Remove(diskPath)
		writeDetail(w, http.StatusInternalServerError, "Could not save metadata")
		return
	}
	msg := map[string]any{
		"type": "attachment_added", "project_id": projectID, "attachment_id": row.ID,
		"post_id": row.PostID, "comment_id": row.CommentID,
	}
	s.emitProject(projectID, msg)
	writeJSON(w, http.StatusOK, s.attachmentMeta(&row))
}

func (s *Server) handleGetAttachment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "attachmentID")
	var row dbpkg.Attachment
	if err := s.DB.First(&row, "id = ?", id).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Attachment not found")
		return
	}
	diskPath := filepath.Join(s.attachmentDir(), row.ID)
	f, err := os.Open(diskPath)
	if err != nil {
		writeDetail(w, http.StatusNotFound, "Attachment data missing")
		return
	}
	defer f.Close()
	disposition := "attachment"
	if strings.HasPrefix(row.ContentType, "image/") || row.ContentType == "application/pdf" {
		disposition = "inline"
	}
	w.Header().Set("Content-Type", row.ContentType)
	w.Header().Set("Content-Disposition", disposition+"; filename=\""+strings.ReplaceAll(row.Filename, "\"", "'")+"\"")
	w.Header().Set("Content-Length", strconv.FormatInt(row.Size, 10))
	http.ServeContent(w, r, row.Filename, row.CreatedAt, f)
}

func (s *Server) handleDeleteAttachment(w http.ResponseWriter, r *http.Request) {
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	id := chi.URLParam(r, "attachmentID")
	var row dbpkg.Attachment
	if err := s.DB.First(&row, "id = ?", id).Error; err != nil {
		writeDetail(w, http.StatusNotFound, "Attachment not found")
		return
	}
	if row.AuthorID != a.ID {
		writeDetail(w, http.StatusForbidden, "Only the uploader can delete this attachment")
		return
	}
	diskPath := filepath.Join(s.attachmentDir(), row.ID)
	_ = os.Remove(diskPath)
	pid := row.ProjectID
	_ = s.DB.Delete(&row).Error
	s.emitProject(pid, map[string]any{
		"type": "attachment_deleted", "project_id": pid, "attachment_id": id,
		"post_id": row.PostID, "comment_id": row.CommentID,
	})
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
