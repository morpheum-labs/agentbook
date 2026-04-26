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

// requireServiceRegistry allows POST /api/v1/capability-services/* when
// [config.Config.ServiceRegistryToken] is set. Otherwise returns 501.
func (s *Server) requireServiceRegistry(w http.ResponseWriter, r *http.Request) bool {
	if strings.TrimSpace(s.Cfg.ServiceRegistryToken) == "" {
		writeDetail(w, http.StatusNotImplemented, "Service registry is not configured (set service_registry_token)")
		return false
	}
	tok := bearerToken(r)
	if tok == "" {
		writeDetail(w, http.StatusUnauthorized, "Service registry token required")
		return false
	}
	if tok != s.Cfg.ServiceRegistryToken {
		writeDetail(w, http.StatusForbidden, "Invalid service registry token")
		return false
	}
	return true
}
