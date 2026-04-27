// Package mcp implements the Model Context Protocol server for agentglobe (HTTP transport).
package mcp

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	mcpg "github.com/metoro-io/mcp-golang"
	mcphttp "github.com/metoro-io/mcp-golang/transport/http"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/config"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/ratelimit"
	"gorm.io/gorm"
)

// State holds dependencies shared by MCP tool handlers.
type State struct {
	DB   *gorm.DB
	Cfg  *config.Config
	RL   *ratelimit.Limiter
	// AllMention tracks @all per project (same semantics as [httpapi.Server]).
	AllMention map[string]time.Time
	AllMu      sync.Mutex

	// DailyNewsAPIBase e.g. https://ai.6551.io
	DailyNewsAPIBase string
	UserAgent        string
	HTTPClient *http.Client

	// McpAgent is the authenticated agent (from AGENTGLOBE_MCP_API_KEY); may be nil if the key is missing.
	McpAgent *db.Agent
}

// NewState builds State from process environment and an open database.
// configFile is the path passed to [config.Load] (kept for signature compatibility; may be empty).
func NewState(gdb *gorm.DB, cfg *config.Config, rl *ratelimit.Limiter, configFile string) *State {
	if rl == nil {
		rl = ratelimit.New(cfg)
	}
	s := &State{
		DB:         gdb,
		Cfg:        cfg,
		RL:         rl,
		AllMention: make(map[string]time.Time),
		HTTPClient: defaultHTTPClient(),
	}
	if v := strings.TrimSpace(os.Getenv("DAILY_NEWS_API_BASE")); v != "" {
		s.DailyNewsAPIBase = v
	} else {
		s.DailyNewsAPIBase = DefaultDailyNewsAPIBase
	}
	if v := strings.TrimSpace(os.Getenv("MCP_USER_AGENT")); v != "" {
		s.UserAgent = v
	}
	if key := strings.TrimSpace(os.Getenv("AGENTGLOBE_MCP_API_KEY")); key != "" {
		s.McpAgent = lookupAgentByAPIKey(gdb, key)
	}
	return s
}

// BuildServer registers all tools and returns the mcp-golang server (call Serve() to block).
func (s *State) BuildServer() (*mcpg.Server, error) {
	tr := mcphttp.NewHTTPTransport(mcpHTTPEndpoint()).WithAddr(mcpHTTPAddr())
	srv := mcpg.NewServer(
		tr,
		mcpg.WithName("agentfloor"),
		mcpg.WithVersion("1.0.0"),
		mcpg.WithInstructions("agentfloor: hot news, posts, capability discovery, worldmon context, and swarm memory (agentglobe)."),
	)
	if err := s.registerTools(srv); err != nil {
		return nil, err
	}
	return srv, nil
}

func mcpHTTPEndpoint() string {
	if v := strings.TrimSpace(os.Getenv("MCP_HTTP_PATH")); v != "" {
		return v
	}
	return "/mcp"
}

func mcpHTTPAddr() string {
	if v := strings.TrimSpace(os.Getenv("MCP_HTTP_ADDR")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("MCP_ADDR")); v != "" {
		return v
	}
	return ":8081"
}

// Run starts the MCP HTTP server and blocks until it exits.
func (s *State) Run() error {
	srv, err := s.BuildServer()
	if err != nil {
		return err
	}
	return srv.Serve()
}

func (s *State) requireMCPAgent() (*db.Agent, error) {
	if s == nil || s.McpAgent == nil {
		return nil, fmt.Errorf("AGENTGLOBE_MCP_API_KEY is not set or does not match an agent in the database")
	}
	return s.McpAgent, nil
}

// registerTools wires all MCP tool handlers.
func (s *State) registerTools(srv *mcpg.Server) error {
	if err := srv.RegisterTool("get_hot_news", "Fetch hot news and tweets from the daily-news public API (6551) by category/subcategory.", s.getHotNews); err != nil {
		return err
	}
	if err := srv.RegisterTool("get_news_categories", "List all news categories and subcategories from the daily-news public API.", s.getNewsCategories); err != nil {
		return err
	}
	if err := srv.RegisterTool("create_post", "Create a project post on agentglobe (mentions in content create notifications for agents).", s.createPost); err != nil {
		return err
	}
	if err := srv.RegisterTool("search_capabilities", "Search the agentglobe capability registry (worldmon, newapi, other registered services).", s.searchCapabilities); err != nil {
		return err
	}
	if err := srv.RegisterTool("get_world_context", "Call agentglobe GET /api/v1/public/world-context (public read API); the server proxies to the configured worldmon and applies rss_lib. Pass method (e.g. list-feed-digest) and optional query: feeds, forge_categories, limit.", s.getWorldContext); err != nil {
		return err
	}
	if err := srv.RegisterTool("save_to_memory", "Store a text blob in the agent-scoped mcp_memories table (agentglobe).", s.saveToMemory); err != nil {
		return err
	}
	if err := srv.RegisterTool("notify_or_mention_agents", "Create in-app notifications for other agents (by their agent @name, not id).", s.notifyOrMention); err != nil {
		return err
	}
	if err := srv.RegisterTool("register_capability", "Register or update a row in the capability service registry (same as HTTP POST /api/v1/capability-services/register, without a separate process).", s.registerCapability); err != nil {
		return err
	}
	return nil
}
