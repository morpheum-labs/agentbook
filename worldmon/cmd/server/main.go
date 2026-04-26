// Command server runs the worldmon HTTP service (World Monitor API proxy, optional agentglobe registration).
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/morpheumlabs/agentbook/worldmon"
	"github.com/morpheumlabs/agentbook/worldmon/internal/httpserver"
	"github.com/morpheumlabs/agentbook/worldmon/internal/regclient"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	cfg := httpserver.LoadConfig()
	key := cfg.WorldMonitorKey
	if key == "" {
		key = strings.TrimSpace(os.Getenv(worldmon.DefaultWorldMonitorKeyEnv))
	}
	var opts []worldmon.Option
	if b := strings.TrimSpace(cfg.WorldMonitorBase); b != "" {
		opts = append(opts, worldmon.WithBaseURL(b))
	} else {
		eb := os.Getenv(worldmon.DefaultWorldMonitorBaseEnv)
		if strings.TrimSpace(eb) != "" {
			opts = append(opts, worldmon.WithBaseURL(eb))
		}
	}
	c := worldmon.New(key, opts...)
	rc := regclient.NewClient(cfg.RegistryBaseURL, cfg.RegistryToken, "morpheumlabs/worldmon")
	if err := httpserver.RunContext(ctx, cfg, c, rc, os.Stderr); err != nil {
		log.Fatal(err)
	}
}
