package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

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
	skill := httpapi.EmbeddedSkill
	if len(skill) == 0 {
		log.Fatal("embedded skill file missing")
	}
	srv := httpapi.NewServer(gdb, cfg, rl, skill, findGitRoot())
	addr := fmt.Sprintf("0.0.0.0:%d", cfg.Port)
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, srv.Handler()))
}
