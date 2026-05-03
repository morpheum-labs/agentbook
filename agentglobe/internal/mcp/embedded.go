package mcp

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/morpheumlabs/agentbook/agentglobe/internal/config"
)

// EmbeddedMCPFromConfig builds an [http.Handler] for POST JSON-RPC MCP at /mcp on the main agentglobe listener.
// Tool handlers use cfg.PublicURL as the agentglobe API origin (same as standalone [StateConfig.GlobeBaseURL]).
func EmbeddedMCPFromConfig(cfg *config.Config) (http.Handler, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	stCfg := &StateConfig{
		GlobeBaseURL:         cfg.PublicURL,
		AgentAPIKey:          strings.TrimSpace(os.Getenv("AGENTGLOBE_MCP_API_KEY")),
		ServiceRegistryToken: firstNonEmpty(strings.TrimSpace(os.Getenv("AGENTGLOBE_SERVICE_REGISTRY_TOKEN")), strings.TrimSpace(cfg.ServiceRegistryToken)),
	}
	if v := strings.TrimSpace(os.Getenv("DAILY_NEWS_API_BASE")); v != "" {
		stCfg.DailyNewsAPIBase = v
	}
	if v := strings.TrimSpace(os.Getenv("MCP_USER_AGENT")); v != "" {
		stCfg.UserAgent = v
	}
	st := NewState(stCfg)
	return st.embeddedHTTPHandler()
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func (s *State) embeddedHTTPHandler() (http.Handler, error) {
	tr := newEmbeddedHTTPTransport()
	srv, err := s.buildServerWithTransport(tr)
	if err != nil {
		return nil, err
	}
	if err := srv.Serve(); err != nil {
		return nil, err
	}
	return tr, nil
}
