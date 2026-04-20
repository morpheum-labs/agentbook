package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"gorm.io/gorm"
)

type connEntry struct {
	c  *websocket.Conn
	mu sync.Mutex
}

// Hub tracks WebSocket connections per agent for realtime fan-out to project members.
type Hub struct {
	mu      sync.RWMutex
	byAgent map[string][]*connEntry
}

func newHub() *Hub {
	return &Hub{byAgent: make(map[string][]*connEntry)}
}

func (h *Hub) register(agentID string, c *websocket.Conn) *connEntry {
	if h == nil {
		return nil
	}
	e := &connEntry{c: c}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.byAgent[agentID] = append(h.byAgent[agentID], e)
	return e
}

func (h *Hub) unregister(agentID string, e *connEntry) {
	if h == nil || e == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	list := h.byAgent[agentID]
	out := list[:0]
	for _, x := range list {
		if x != e {
			out = append(out, x)
		}
	}
	if len(out) == 0 {
		delete(h.byAgent, agentID)
	} else {
		h.byAgent[agentID] = out
	}
}

func (h *Hub) broadcastAll(msg map[string]any) {
	if h == nil {
		return
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, list := range h.byAgent {
		for _, e := range list {
			e.mu.Lock()
			_ = e.c.WriteMessage(websocket.TextMessage, b)
			e.mu.Unlock()
		}
	}
}

func (h *Hub) broadcastToProjectMembers(db *gorm.DB, projectID string, msg map[string]any) {
	if h == nil {
		return
	}
	var ids []string
	if err := db.Model(&dbpkg.ProjectMember{}).Where("project_id = ?", projectID).Pluck("agent_id", &ids).Error; err != nil || len(ids) == 0 {
		return
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, aid := range ids {
		for _, e := range h.byAgent[aid] {
			e.mu.Lock()
			_ = e.c.WriteMessage(websocket.TextMessage, b)
			e.mu.Unlock()
		}
	}
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     s.wsCheckOrigin,
	}
	token := r.URL.Query().Get("token")
	if token == "" {
		writeDetail(w, http.StatusUnauthorized, "Missing token query parameter (agent API key)")
		return
	}
	var a dbpkg.Agent
	if err := s.dbCtx(r).Where("api_key = ?", token).First(&a).Error; err != nil {
		writeDetail(w, http.StatusUnauthorized, "Invalid token")
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	entry := s.Hub.register(a.ID, conn)
	_ = conn.WriteJSON(map[string]any{"type": "connected", "agent_id": a.ID})
	go func() {
		defer func() {
			s.Hub.unregister(a.ID, entry)
			_ = conn.Close()
		}()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
}

func (s *Server) emitProject(projectID string, msg map[string]any) {
	if s.Hub == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	s.Hub.broadcastToProjectMembers(s.DB.WithContext(ctx), projectID, msg)
}

// emitFloor broadcasts a JSON text frame to every connected WebSocket client (chamber / live session channel; V3 name).
func (s *Server) emitFloor(msg map[string]any) {
	if s.Hub == nil {
		return
	}
	s.Hub.broadcastAll(msg)
}
