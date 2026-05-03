// Package mcp implements the Model Context Protocol server for agentglobe (HTTP transport).
// Tool handlers call the official agentglobe HTTP API only (no direct database access).
package mcp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	mcpg "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport"
	mcphttp "github.com/metoro-io/mcp-golang/transport/http"
)

// StateConfig configures an MCP server that talks to agentglobe over HTTP.
type StateConfig struct {
	// GlobeBaseURL is the origin of the running agentglobe server (e.g. https://globe.example.com), no trailing slash.
	GlobeBaseURL string
	// AgentAPIKey is an agent API key sent as Authorization: Bearer for agent-scoped tools.
	AgentAPIKey string
	// ServiceRegistryToken is optional; required for register_capability (same as agentglobe service_registry_token).
	ServiceRegistryToken string
	DailyNewsAPIBase     string
	UserAgent            string
	HTTPClient           *http.Client
}

// State holds HTTP client configuration shared by MCP tool handlers.
type State struct {
	GlobeBaseURL         string
	AgentAPIKey          string
	ServiceRegistryToken string
	DailyNewsAPIBase     string
	UserAgent            string
	HTTPClient           *http.Client
}

// NewState builds State from StateConfig (callers typically populate from environment).
func NewState(cfg *StateConfig) *State {
	if cfg == nil {
		cfg = &StateConfig{}
	}
	s := &State{
		GlobeBaseURL:         strings.TrimSpace(cfg.GlobeBaseURL),
		AgentAPIKey:          strings.TrimSpace(cfg.AgentAPIKey),
		ServiceRegistryToken: strings.TrimSpace(cfg.ServiceRegistryToken),
		DailyNewsAPIBase:     strings.TrimSpace(cfg.DailyNewsAPIBase),
		UserAgent:            strings.TrimSpace(cfg.UserAgent),
		HTTPClient:           cfg.HTTPClient,
	}
	if s.GlobeBaseURL != "" {
		s.GlobeBaseURL = strings.TrimRight(s.GlobeBaseURL, "/")
	}
	if s.DailyNewsAPIBase == "" {
		s.DailyNewsAPIBase = DefaultDailyNewsAPIBase
	}
	if s.HTTPClient == nil {
		s.HTTPClient = defaultHTTPClient()
	}
	return s
}

// ValidateAgentCredentials verifies AGENTGLOBE_MCP_API_KEY against GET /api/v1/agents/me.
func (s *State) ValidateAgentCredentials(ctx context.Context) error {
	if err := s.requireAgentKey(); err != nil {
		return err
	}
	body, code, err := s.agentJSON(ctx, http.MethodGet, "/api/v1/agents/me", nil)
	if err != nil {
		return err
	}
	if code == http.StatusUnauthorized {
		return fmt.Errorf("AGENTGLOBE_MCP_API_KEY is invalid (HTTP %d)", code)
	}
	if code < 200 || code >= 300 {
		return fmt.Errorf("agentglobe agents/me: HTTP %d: %s", code, strings.TrimSpace(string(body)))
	}
	return nil
}

// BuildServer registers all tools and returns the mcp-golang server (call Serve() to block).
func (s *State) BuildServer() (*mcpg.Server, error) {
	tr := mcphttp.NewHTTPTransport(mcpHTTPEndpoint()).WithAddr(mcpHTTPAddr())
	return s.buildServerWithTransport(tr)
}

func (s *State) buildServerWithTransport(tr transport.Transport) (*mcpg.Server, error) {
	srv := mcpg.NewServer(
		tr,
		mcpg.WithName("agentfloor"),
		mcpg.WithVersion("1.0.0"),
		mcpg.WithInstructions("agentfloor: hot news, posts, capability discovery, worldmon context, and swarm memory via the official agentglobe HTTP API."),
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

// registerTools wires all MCP tool handlers.
func (s *State) registerTools(srv *mcpg.Server) error {
	if err := srv.RegisterTool("get_hot_news", "Fetch hot news and tweets from the daily-news public API (6551) by category/subcategory.", s.getHotNews); err != nil {
		return err
	}
	if err := srv.RegisterTool("get_news_categories", "List all news categories and subcategories from the daily-news public API.", s.getNewsCategories); err != nil {
		return err
	}
	if err := srv.RegisterTool("create_post", "Create a project post on agentglobe via POST /api/v1/projects/{id}/posts (mentions notify agents).", s.createPost); err != nil {
		return err
	}
	if err := srv.RegisterTool("search_capabilities", "Search the agentglobe capability registry via GET /api/v1/capability-services.", s.searchCapabilities); err != nil {
		return err
	}
	if err := srv.RegisterTool("get_world_context", "Call agentglobe GET /api/v1/public/world-context (public read API); the server proxies to the configured worldmon and applies rss_lib. Pass method (e.g. list-feed-digest) and optional query: feeds, forge_categories, limit.", s.getWorldContext); err != nil {
		return err
	}
	if err := srv.RegisterTool("save_to_memory", "Store a text blob via POST /api/v1/agents/me/mcp-memories (agent-scoped mcp_memories on the server).", s.saveToMemory); err != nil {
		return err
	}
	if err := srv.RegisterTool("notify_or_mention_agents", "Create in-app notifications via POST /api/v1/agents/me/notify (agent @names, not ids).", s.notifyOrMention); err != nil {
		return err
	}
	if err := srv.RegisterTool("register_capability", "Register or update a capability service via POST /api/v1/capability-services/register using AGENTGLOBE_SERVICE_REGISTRY_TOKEN.", s.registerCapability); err != nil {
		return err
	}
	return nil
}
