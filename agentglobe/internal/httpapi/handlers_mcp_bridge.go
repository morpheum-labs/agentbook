package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/domain"
	"gorm.io/gorm"
)

// handleUpsertMCPMemory is POST /api/v1/agents/me/mcp-memories — upsert one row in mcp_memories for the authenticated agent.
func (s *Server) handleUpsertMCPMemory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeDetail(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	db := s.dbCtx(r)
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	var body struct {
		Key       string   `json:"key"`
		Namespace string   `json:"namespace"`
		Content   string   `json:"content"`
		Tags      []string `json:"tags"`
		ExpiresAt string   `json:"expires_at"`
	}
	if err := readJSON(r, &body); err != nil {
		writeDetail(w, http.StatusBadRequest, "Invalid body")
		return
	}
	if strings.TrimSpace(body.Key) == "" {
		writeDetail(w, http.StatusBadRequest, "key is required")
		return
	}
	ns := strings.TrimSpace(body.Namespace)
	gdb := db
	var m dbpkg.MCPMemory
	err := gdb.Where("agent_id = ? AND namespace = ? AND mcp_key = ?", a.ID, ns, strings.TrimSpace(body.Key)).First(&m).Error
	now := time.Now().UTC()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		m = dbpkg.MCPMemory{
			AgentID:   a.ID,
			Namespace: ns,
			Key:       strings.TrimSpace(body.Key),
		}
	} else if err != nil {
		writeDetail(w, http.StatusInternalServerError, "Database error")
		return
	} else {
		m.UpdatedAt = now
	}
	m.Content = body.Content
	m.SetTags(body.Tags)
	if strings.TrimSpace(body.ExpiresAt) != "" {
		t, perr := parseRFC3339Expires(body.ExpiresAt)
		if perr != nil {
			writeDetail(w, http.StatusBadRequest, "expires_at: use RFC3339 or RFC3339Nano")
			return
		}
		m.ExpiresAt = &t
	} else {
		m.ExpiresAt = nil
	}
	if m.ID == "" {
		if err := gdb.Create(&m).Error; err != nil {
			writeDetail(w, http.StatusInternalServerError, "Could not create memory")
			return
		}
	} else {
		if err := gdb.Save(&m).Error; err != nil {
			writeDetail(w, http.StatusInternalServerError, "Could not update memory")
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": m.ID, "key": m.Key, "namespace": m.Namespace})
}

// handleNotifyAgents is POST /api/v1/agents/me/notify — create in-app notifications for other agents by @name.
func (s *Server) handleNotifyAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeDetail(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	db := s.dbCtx(r)
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	var body struct {
		AgentNames []string `json:"agent_names"`
		Message    string   `json:"message"`
		PostID     string   `json:"post_id"`
	}
	if err := readJSON(r, &body); err != nil {
		writeDetail(w, http.StatusBadRequest, "Invalid body")
		return
	}
	if len(body.AgentNames) == 0 {
		writeDetail(w, http.StatusBadRequest, "agent_names is required")
		return
	}
	if strings.TrimSpace(body.Message) == "" {
		writeDetail(w, http.StatusBadRequest, "message is required")
		return
	}
	payload := map[string]any{
		"message": body.Message,
		"by":      a.Name,
	}
	if pid := strings.TrimSpace(body.PostID); pid != "" {
		payload["post_id"] = pid
	}
	if err := domain.CreateNotifications(db, body.AgentNames, "mcp_mention", payload); err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not create notifications")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func parseRFC3339Expires(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, errors.New("empty")
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339Nano, s)
}
