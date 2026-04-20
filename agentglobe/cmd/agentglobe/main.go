package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/morpheumlabs/agentbook/agentglobe/internal/config"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/httpapi"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/ratelimit"
)

func findGitRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func envDuration(key string, def time.Duration) time.Duration {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return def
	}
	if s == "0" {
		return 0
	}
	d, err := time.ParseDuration(s)
	if err != nil || d < 0 {
		return def
	}
	return d
}

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
	if strings.EqualFold(strings.TrimSpace(os.Getenv("AGENTGLOBE_FLOOR_SEED_DEMO")), "1") {
		if err := db.SeedFloorDemoTopics(gdb); err != nil {
			log.Printf("AGENTGLOBE_FLOOR_SEED_DEMO: seed failed: %v", err)
		}
	}
	rl := ratelimit.New(cfg)
	skill := httpapi.EmbeddedSkill
	if len(skill) == 0 {
		log.Fatal("embedded skill file missing")
	}
	srv := httpapi.NewServer(gdb, cfg, rl, skill, findGitRoot())
	addr := fmt.Sprintf("0.0.0.0:%d", cfg.Port)
	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: envDuration("HTTP_READ_HEADER_TIMEOUT", 10*time.Second),
		ReadTimeout:       envDuration("HTTP_READ_TIMEOUT", 10*time.Minute),
		WriteTimeout:      envDuration("HTTP_WRITE_TIMEOUT", 10*time.Minute),
		IdleTimeout:       envDuration("HTTP_IDLE_TIMEOUT", 3*time.Minute),
	}
	log.Printf("listening on %s", addr)
	log.Fatal(httpSrv.ListenAndServe())
}
