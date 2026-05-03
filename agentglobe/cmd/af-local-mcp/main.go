// Command af-local-mcp runs the Model Context Protocol HTTP server for agentglobe.
// It calls the official agentglobe HTTP API only (AGENTGLOBE_BASE_URL); it does not open the database or load server config files.
package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/morpheumlabs/agentbook/agentglobe/internal/mcp"
)

func main() {
	base := strings.TrimSpace(os.Getenv("AGENTGLOBE_BASE_URL"))
	if base == "" {
		log.Fatal("AGENTGLOBE_BASE_URL is required (origin of the running agentglobe server, e.g. https://globe.example.com)")
	}
	key := strings.TrimSpace(os.Getenv("AGENTGLOBE_MCP_API_KEY"))
	cfg := &mcp.StateConfig{
		GlobeBaseURL:         base,
		AgentAPIKey:          key,
		ServiceRegistryToken: strings.TrimSpace(os.Getenv("AGENTGLOBE_SERVICE_REGISTRY_TOKEN")),
	}
	if v := strings.TrimSpace(os.Getenv("DAILY_NEWS_API_BASE")); v != "" {
		cfg.DailyNewsAPIBase = v
	}
	if v := strings.TrimSpace(os.Getenv("MCP_USER_AGENT")); v != "" {
		cfg.UserAgent = v
	}
	st := mcp.NewState(cfg)
	if key == "" {
		if strings.TrimSpace(os.Getenv("AGENTGLOBE_MCP_STRICT")) == "1" {
			log.Fatal("AGENTGLOBE_MCP_STRICT: set AGENTGLOBE_MCP_API_KEY to a valid agent API key")
		}
		log.Print("warning: AGENTGLOBE_MCP_API_KEY not set; create_post, save_to_memory, and notify_or_mention_agents will fail until configured")
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		err := st.ValidateAgentCredentials(ctx)
		cancel()
		if err != nil {
			if strings.TrimSpace(os.Getenv("AGENTGLOBE_MCP_STRICT")) == "1" {
				log.Fatalf("AGENTGLOBE_MCP_STRICT: %v", err)
			}
			log.Printf("warning: could not validate AGENTGLOBE_MCP_API_KEY against agentglobe: %v", err)
		}
	}
	if err := st.Run(); err != nil {
		log.Fatal(err)
	}
}
