package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/db"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/events"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/httperr"
	"gorm.io/gorm"
)

// RouterOptions configures optional HTTP router behavior.
type RouterOptions struct {
	InternalToken string // When set, POST /api/v1/events/publish requires Bearer or X-Internal-Token.
	APIKey        string // When set, /api/v1/* requires Bearer or X-API-Key (healthz, openapi, metrics exempt).
	RateLimitRPS  float64 // When > 0, rate-limit /api/v1/* per client IP (sustained RPS).
	MaxBodyBytes  int64   // Max JSON body size for /api/v1/* (default 1 MiB when 0).
	// CredentialsMasterKey is a 32-byte AES-256 key for encrypting agent credentials at rest (optional).
	CredentialsMasterKey []byte
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
	maxBody := opts.MaxBodyBytes
	if maxBody <= 0 {
		maxBody = 1 << 20
	}
	s := &Server{
		db:             gdb,
		hub:            events.NewHub(),
		dispatcher:     &events.WebhookDispatcher{DB: gdb, Client: events.DefaultHTTPClient()},
		internalToken:  opts.InternalToken,
		apiKey:         opts.APIKey,
		maxBodyBytes:   maxBody,
		credMasterKey:  append([]byte(nil), opts.CredentialsMasterKey...),
	}
	r := chi.NewRouter()
	r.Use(corsMiddleware)
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(slogRequestLogger)
	r.Get("/healthz", s.healthz)
	r.Get("/openapi.json", handleOpenapi())
	r.Handle("/metrics", promhttp.Handler())

	var rl *ipRateLimiter
	if opts.RateLimitRPS > 0 {
		burst := int(opts.RateLimitRPS) + 5
		if burst < 5 {
			burst = 5
		}
		rl = newIPRateLimiter(opts.RateLimitRPS, burst, 15*time.Minute)
		go rl.cleanupLoop(context.Background())
	}

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(maxBodyBytes(maxBody))
		if rl != nil {
			r.Use(rl.middleware)
		}
		r.Use(s.requireAPIKey)
		r.Get("/config", s.getConfig)
		r.Put("/config", s.putConfig)

		r.Get("/agents", s.listAgents)
		r.Post("/agents", s.createAgent)
		r.Get("/agents/{id}", s.getAgent)
		r.Put("/agents/{id}", s.putAgent)
		r.Patch("/agents/{id}", s.patchAgent)
		r.Delete("/agents/{id}", s.deleteAgent)
		r.Get("/agents/by-name/{name}", s.getAgentByName)

		r.Get("/agents/{id}/credentials", s.listAgentCredentials)
		r.Post("/agents/{id}/credentials", s.createAgentCredential)
		r.Post("/agents/{id}/credentials/{bindingId}/rotate", s.rotateAgentCredential)
		r.Delete("/agents/{id}/credentials/{bindingId}", s.deleteAgentCredential)

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
	db             *gorm.DB
	hub            *events.Hub
	dispatcher     *events.WebhookDispatcher
	internalToken  string
	apiKey         string
	maxBodyBytes   int64
	credMasterKey  []byte
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{
		"status": "ok",
		"ts":     time.Now().UTC().Format(time.RFC3339Nano),
	}
	code := http.StatusOK
	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err != nil {
			payload["status"] = "degraded"
			payload["database"] = map[string]string{"error": err.Error()}
			code = http.StatusServiceUnavailable
		} else if err := sqlDB.PingContext(r.Context()); err != nil {
			payload["status"] = "degraded"
			payload["database"] = map[string]string{"error": err.Error()}
			code = http.StatusServiceUnavailable
		} else {
			payload["database"] = "ok"
			var n int64
			_ = s.db.Model(&db.SwarmRuntimeInstance{}).Count(&n)
			payload["runtime_instances"] = n
		}
	} else {
		payload["database"] = "not_configured"
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
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
