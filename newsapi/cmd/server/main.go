// Command server runs the newsapi HTTP service (health, OpenAPI, News API proxy, agentglobe registration).
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/morpheumlabs/agentbook/newsapi/internal/httpserver"
	"github.com/morpheumlabs/agentbook/newsapi/internal/regclient"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	cfg := httpserver.LoadConfig()
	rc := regclient.NewClient(cfg.RegistryBaseURL, cfg.RegistryToken, "morpheumlabs/newsapi")
	if err := httpserver.RunContext(ctx, cfg, rc, os.Stderr); err != nil {
		log.Fatal(err)
	}
}
