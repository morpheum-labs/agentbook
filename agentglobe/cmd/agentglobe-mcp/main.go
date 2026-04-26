// Command agentglobe-mcp runs the Model Context Protocol HTTP server for agentglobe.
package main

import (
	"log"
	"os"
	"strings"

	"github.com/morpheumlabs/agentbook/agentglobe/internal/config"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/mcp"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/ratelimit"
)

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = config.DefaultConfigPath()
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	gdb, err := db.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	rl := ratelimit.New(cfg)
	st := mcp.NewState(gdb, cfg, rl, cfgPath)
	if st.McpAgent == nil {
		if strings.TrimSpace(os.Getenv("AGENTGLOBE_MCP_STRICT")) == "1" {
			log.Fatal("AGENTGLOBE_MCP_STRICT: set AGENTGLOBE_MCP_API_KEY to a valid agent API key")
		}
		log.Print("warning: AGENTGLOBE_MCP_API_KEY not set or not found; create_post, save_to_memory, and notify_or_mention_agents will fail until configured")
	}
	if err := st.Run(); err != nil {
		log.Fatal(err)
	}
}
