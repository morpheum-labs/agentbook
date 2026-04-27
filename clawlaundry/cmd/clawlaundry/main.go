// Command clawlaundry is the HTTP API for MiroClaw/ZeroClaw-style swarm agent metadata
// (Hands + cron jobs + defaults) backed by PostgreSQL and GORM.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/morpheumlabs/agentbook/clawlaundry/internal/api"
	"github.com/morpheumlabs/agentbook/clawlaundry/internal/config"
	"github.com/morpheumlabs/agentbook/clawlaundry/internal/db"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "YAML config (e.g. dep/cl.yaml); same as -c; default: CONFIG_PATH or first match from DefaultConfigPath")
	flag.StringVar(&configPath, "c", "", "YAML config (shorthand for -config)")
	flag.Parse()
	if configPath == "" {
		configPath = strings.TrimSpace(os.Getenv("CONFIG_PATH"))
	}
	if configPath == "" {
		configPath = config.DefaultConfigPath()
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("config", "err", err, "path", configPath)
		os.Exit(1)
	}
	if cfg.DatabaseURL == "" {
		slog.Error("database is required: set database_url in YAML (e.g. dep/cl.yaml) or DATABASE_URL")
		os.Exit(1)
	}
	gormDB, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		slog.Error("db open", "err", err)
		os.Exit(1)
	}

	h := api.NewRouter(gormDB)
	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: h,
	}
	slog.Info("clawlaundry listening", "addr", cfg.HTTPAddr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server", "err", err)
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
