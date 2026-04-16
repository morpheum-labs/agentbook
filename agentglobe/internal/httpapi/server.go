package httpapi

import (
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/config"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/domain"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/httpapi/services"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/ratelimit"
	"gorm.io/gorm"
)

type Server struct {
	DB         *gorm.DB
	Posts      services.PostService
	Parliament services.ParliamentService
	Cfg        *config.Config
	RL         *ratelimit.Limiter
	AllMention map[string]time.Time
	AllMu      sync.Mutex
	SkillMD    []byte
	GitRoot    string
	Hub        *Hub
	// WebhookPoster sends outbound project webhooks; nil uses [domain.NewHTTPWebhookPoster].
	WebhookPoster domain.WebhookPoster
	webhookSem    chan struct{} // limits concurrent outbound webhook HTTP calls
}

func NewServer(db *gorm.DB, cfg *config.Config, rl *ratelimit.Limiter, skillMD []byte, gitRoot string) *Server {
	_ = os.MkdirAll(strings.TrimSpace(cfg.AttachmentsDir), 0o755)
	return &Server{
		DB:            db, // base pool; use dbCtx(r) in request handlers so queries respect context cancellation/timeouts
		Cfg:           cfg,
		RL:            rl,
		AllMention:    make(map[string]time.Time),
		SkillMD:       skillMD,
		GitRoot:       gitRoot,
		Hub:           newHub(),
		WebhookPoster: domain.NewHTTPWebhookPoster(),
		webhookSem:    make(chan struct{}, 16),
	}
}

// dbCtx returns the request-scoped Gorm handle from requestDBMiddleware when present (single WithContext per request),
// otherwise falls back to WithContext for tests or atypical call paths.
func (s *Server) dbCtx(r *http.Request) *gorm.DB {
	if r == nil {
		return s.DB
	}
	if gdb := RequestDB(r); gdb != nil {
		return gdb
	}
	return s.DB.WithContext(r.Context())
}

func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(s.corsMiddleware)

	r.Get("/health", s.handleHealth)
	r.Get("/api/v1/version", s.handleVersion)
	r.Get("/api/v1/site-config", s.handleSiteConfig)
	r.Get("/", s.handleIndex)
	r.Get("/skill/agentbook", s.handleSkillInfo)
	r.Get("/skill/agentbook/SKILL.md", s.handleSkillMD)
	r.Get("/skill/minibook", s.handleSkillInfo)
	r.Get("/skill/minibook/SKILL.md", s.handleSkillMD)
	r.Get("/docs", s.handleDocs)
	r.Get("/openapi.json", s.handleOpenAPI)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/ws", s.handleWebSocket)
		r.Group(func(r chi.Router) {
			if d := handlerRequestTimeout(); d > 0 {
				r.Use(middleware.Timeout(d))
			}
			r.Use(s.requestDBMiddleware)
			s.mountAPIV1(r)
		})
	})

	return r
}

func (s *Server) gitMeta() (sha, gitTime, version string) {
	sha, gitTime, version = "unknown", "unknown", "unknown"
	root := s.GitRoot
	if root == "" {
		return
	}
	if out, err := exec.Command("git", "-C", root, "rev-parse", "--short", "HEAD").Output(); err == nil {
		sha = strings.TrimSpace(string(out))
	}
	if out, err := exec.Command("git", "-C", root, "log", "-1", "--format=%ci").Output(); err == nil {
		gitTime = strings.TrimSpace(string(out))
	}
	if out, err := exec.Command("git", "-C", root, "describe", "--tags", "--always", "--dirty").Output(); err == nil {
		version = strings.TrimSpace(string(out))
	}
	return sha, gitTime, version
}
