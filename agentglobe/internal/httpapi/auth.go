package httpapi

import (
	"net/http"
	"strings"

	"github.com/morpheumlabs/agentbook/agentglobe/internal/db"
)

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	const p = "Bearer "
	if len(h) > len(p) && strings.EqualFold(h[:len(p)], p) {
		return strings.TrimSpace(h[len(p):])
	}
	return ""
}

func (s *Server) currentAgent(r *http.Request) *db.Agent {
	key := bearerToken(r)
	if key == "" {
		return nil
	}
	var a db.Agent
	if err := s.dbCtx(r).Where("api_key = ?", key).First(&a).Error; err != nil {
		return nil
	}
	return &a
}

func (s *Server) requireAgent(w http.ResponseWriter, r *http.Request) *db.Agent {
	a := s.currentAgent(r)
	if a == nil {
		writeDetail(w, http.StatusUnauthorized, "Invalid or missing API key")
		return nil
	}
	return a
}

func (s *Server) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	if strings.TrimSpace(s.Cfg.AdminToken) == "" {
		writeDetail(w, http.StatusInternalServerError, "Admin token not configured")
		return false
	}
	tok := bearerToken(r)
	if tok == "" {
		writeDetail(w, http.StatusUnauthorized, "Admin token required")
		return false
	}
	if tok != s.Cfg.AdminToken {
		writeDetail(w, http.StatusForbidden, "Invalid admin token")
		return false
	}
	return true
}
