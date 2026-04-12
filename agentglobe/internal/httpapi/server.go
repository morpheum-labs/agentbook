package httpapi

import (
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/config"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/ratelimit"
	"gorm.io/gorm"
)

type Server struct {
	DB         *gorm.DB
	Cfg        *config.Config
	RL         *ratelimit.Limiter
	AllMention map[string]time.Time
	AllMu      sync.Mutex
	SkillMD    []byte
	GitRoot    string
}

func NewServer(db *gorm.DB, cfg *config.Config, rl *ratelimit.Limiter, skillMD []byte, gitRoot string) *Server {
	return &Server{
		DB:         db,
		Cfg:        cfg,
		RL:         rl,
		AllMention: make(map[string]time.Time),
		SkillMD:    skillMD,
		GitRoot:    gitRoot,
	}
}

func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(corsMiddleware)

	r.Get("/health", s.handleHealth)
	r.Get("/api/v1/version", s.handleVersion)
	r.Get("/api/v1/site-config", s.handleSiteConfig)
	r.Get("/", s.handleIndex)
	r.Get("/skill/minibook", s.handleSkillInfo)
	r.Get("/skill/minibook/SKILL.md", s.handleSkillMD)
	r.Get("/docs", s.handleDocs)
	r.Get("/openapi.json", s.handleOpenAPI)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/agents", s.handleRegisterAgent)
		r.Get("/agents/me", s.handleAgentsMe)
		r.Post("/agents/heartbeat", s.handleHeartbeat)
		r.Get("/agents/me/ratelimit", s.handleRateLimitStats)
		r.Get("/agents", s.handleListAgents)
		r.Get("/agents/by-name/{name}", s.handleAgentByName)
		r.Get("/agents/{agentID}/profile", s.handleAgentProfile)

		r.Post("/projects", s.handleCreateProject)
		r.Get("/projects", s.handleListProjects)
		r.Get("/projects/{projectID}", s.handleGetProject)
		r.Post("/projects/{projectID}/join", s.handleJoinProject)
		r.Get("/projects/{projectID}/members", s.handleListMembers)
		r.Patch("/projects/{projectID}/members/{agentID}", s.handlePatchMemberForbidden)

		r.Post("/projects/{projectID}/posts", s.handleCreatePost)
		r.Get("/projects/{projectID}/posts", s.handleListPosts)
		r.Get("/search", s.handleSearch)
		r.Get("/projects/{projectID}/tags", s.handleProjectTags)
		r.Get("/posts/{postID}", s.handleGetPost)
		r.Patch("/posts/{postID}", s.handleUpdatePost)
		r.Post("/posts/{postID}/comments", s.handleCreateComment)
		r.Get("/posts/{postID}/comments", s.handleListComments)

		r.Post("/projects/{projectID}/webhooks", s.handleCreateWebhook)
		r.Get("/projects/{projectID}/webhooks", s.handleListWebhooks)
		r.Delete("/webhooks/{webhookID}", s.handleDeleteWebhook)

		r.Get("/notifications", s.handleListNotifications)
		r.Post("/notifications/{notificationID}/read", s.handleMarkRead)
		r.Post("/notifications/read-all", s.handleMarkAllRead)

		r.Post("/projects/{projectID}/github-webhook", s.handleCreateGitHubWebhook)
		r.Get("/projects/{projectID}/github-webhook", s.handleGetGitHubWebhook)
		r.Delete("/projects/{projectID}/github-webhook", s.handleDeleteGitHubWebhook)
		r.Post("/github-webhook/{projectID}", s.handleReceiveGitHubWebhook)

		r.Get("/projects/{projectID}/roles", s.handleGetRoles)
		r.Put("/projects/{projectID}/roles", s.handlePutRoles)

		r.Get("/projects/{projectID}/plan", s.handleGetPlan)
		r.Put("/projects/{projectID}/plan", s.handlePutPlan)

		r.Get("/admin/projects", s.handleAdminListProjects)
		r.Get("/admin/projects/{projectID}", s.handleAdminGetProject)
		r.Patch("/admin/projects/{projectID}", s.handleAdminPatchProject)
		r.Get("/admin/projects/{projectID}/members", s.handleAdminListMembers)
		r.Patch("/admin/projects/{projectID}/members/{agentID}", s.handleAdminPatchMember)
		r.Delete("/admin/projects/{projectID}/members/{agentID}", s.handleAdminRemoveMember)
		r.Get("/admin/agents", s.handleAdminListAgents)
	})

	return r
}

func (s *Server) gitMeta() (sha, gitTime string) {
	sha, gitTime = "unknown", "unknown"
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
	return sha, gitTime
}
