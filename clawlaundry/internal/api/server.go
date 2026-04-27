package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/morpheumlabs/agentbook/clawlaundry/internal/db"
	"github.com/morpheumlabs/agentbook/clawlaundry/internal/httperr"
	"gorm.io/gorm"
)

// corsMiddleware allows browser UIs (e.g. a static SPA on another origin) to call the JSON API.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// NewRouter mounts REST handlers on r (caller may wrap with middleware).
func NewRouter(gdb *gorm.DB) http.Handler {
	s := &Server{db: gdb}
	r := chi.NewRouter()
	r.Use(corsMiddleware)
	r.Get("/healthz", s.healthz)
	r.Get("/openapi.json", handleOpenapi())

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/config", s.getConfig)
		r.Put("/config", s.putConfig)

		r.Get("/agents", s.listAgents)
		r.Post("/agents", s.createAgent)
		r.Get("/agents/{id}", s.getAgent)
		r.Put("/agents/{id}", s.putAgent)
		r.Patch("/agents/{id}", s.patchAgent)
		r.Delete("/agents/{id}", s.deleteAgent)
		// Optional convenience: resolve by hand name
		r.Get("/agents/by-name/{name}", s.getAgentByName)

		r.Get("/cron-jobs", s.listCronJobs)
		r.Post("/cron-jobs", s.createCronJob)
		r.Get("/cron-jobs/{id}", s.getCronJob)
		r.Put("/cron-jobs/{id}", s.putCronJob)
		r.Patch("/cron-jobs/{id}", s.patchCronJob)
		r.Delete("/cron-jobs/{id}", s.deleteCronJob)
	})
	return r
}

// Server holds shared dependencies.
type Server struct {
	db *gorm.DB
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "ts": time.Now().UTC().Format(time.RFC3339Nano)})
}

func (s *Server) getConfig(w http.ResponseWriter, r *http.Request) {
	var c db.SwarmConfig
	if err := s.db.First(&c, 1).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

type putConfigBody struct {
	DefaultProvider string `json:"default_provider"`
	DefaultModel    string `json:"default_model"`
}

func (s *Server) putConfig(w http.ResponseWriter, r *http.Request) {
	var b putConfigBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid json body", err))
		return
	}
	c := db.SwarmConfig{ID: 1, DefaultProvider: b.DefaultProvider, DefaultModel: b.DefaultModel}
	if err := s.db.Save(&c).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}
