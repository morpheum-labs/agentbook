package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/db"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/events"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/httperr"
	"gorm.io/gorm"
)

// RouterOptions configures optional HTTP router behavior.
type RouterOptions struct {
	InternalToken string // When set, POST /api/v1/events/publish requires Bearer or X-Internal-Token.
}

// corsMiddleware allows browser UIs (e.g. a static SPA on another origin) to call the JSON API.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if v := r.Header.Get("Access-Control-Request-Headers"); v != "" {
			w.Header().Set("Access-Control-Allow-Headers", v)
		} else {
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Authorization")
		}
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// NewRouter mounts REST handlers on r (caller may wrap with middleware).
func NewRouter(gdb *gorm.DB, opts RouterOptions) http.Handler {
	s := &Server{
		db:            gdb,
		hub:           events.NewHub(),
		dispatcher:    &events.WebhookDispatcher{DB: gdb, Client: events.DefaultHTTPClient()},
		internalToken: opts.InternalToken,
	}
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
		r.Get("/agents/by-name/{name}", s.getAgentByName)

		r.Get("/cron-jobs/schedule-timeline", s.listCronScheduleTimeline)
		r.Get("/cron-jobs", s.listCronJobs)
		r.Post("/cron-jobs", s.createCronJob)
		r.Get("/cron-jobs/{id}", s.getCronJob)
		r.Put("/cron-jobs/{id}", s.putCronJob)
		r.Patch("/cron-jobs/{id}", s.patchCronJob)
		r.Delete("/cron-jobs/{id}", s.deleteCronJob)

		r.Post("/instances/register", s.registerInstance)
		r.Post("/instances/{instance_name}/heartbeat", s.heartbeatInstance)
		r.Get("/instances", s.listInstances)
		r.Get("/instances/{instance_name}", s.getInstance)
		r.Delete("/instances/{instance_name}", s.deleteInstance)

		r.With(s.requireInternalToken).Post("/events/publish", s.publishEvent)
		r.Get("/events", s.streamEvents)
	})
	return r
}

// Server holds shared dependencies.
type Server struct {
	db            *gorm.DB
	hub           *events.Hub
	dispatcher    *events.WebhookDispatcher
	internalToken string
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
	sum, err := db.LoadRevisionSummary(s.db)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"default_provider": c.DefaultProvider,
		"default_model":    c.DefaultModel,
		"current_revision": c.CurrentRevision,
		"revision_summary": sum,
	})
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
	now := time.Now().UTC()
	if err := s.db.Model(&db.SwarmConfig{}).Where("id = ?", 1).Updates(map[string]any{
		"default_provider": b.DefaultProvider,
		"default_model":    b.DefaultModel,
		"current_revision": gorm.Expr("current_revision + 1"),
		"updated_at":       now,
	}).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	var c db.SwarmConfig
	if err := s.db.First(&c, 1).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	s.emit(events.ChangeEvent{
		EventType:          events.EventConfigUpdated,
		AffectedEntityType: events.EntityConfig,
		AffectedIDs:        []string{"1"},
		NewRevision:        c.CurrentRevision,
		TS:                 events.NowRFC3339Nano(),
	})
	sum, err := db.LoadRevisionSummary(s.db)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"default_provider": c.DefaultProvider,
		"default_model":    c.DefaultModel,
		"current_revision": c.CurrentRevision,
		"revision_summary": sum,
	})
}
