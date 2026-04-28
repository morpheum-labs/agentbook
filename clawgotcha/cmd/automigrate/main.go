// Command automigrate connects to PostgreSQL and runs GORM AutoMigrate (same as server startup).
// Config resolution matches clawgotcha: -c / CONFIG_PATH / dep/cl.yaml (see config.DefaultConfigPath).
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/morpheumlabs/agentbook/clawgotcha/internal/config"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/db"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "c", "", "YAML config (e.g. dep/cl.yaml); default: CONFIG_PATH or search dep/cl.yaml, ../dep/cl.yaml")
	flag.Parse()

	p := strings.TrimSpace(configPath)
	if p == "" {
		p = strings.TrimSpace(os.Getenv("CONFIG_PATH"))
	}
	if p == "" {
		p = config.DefaultConfigPath()
	}

	cfg, err := config.Load(p)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if p != "" {
		fmt.Fprintf(os.Stderr, "config: %s\n", p)
	}
	if cfg.DatabaseURL == "" {
		fmt.Fprintln(os.Stderr, "database_url is empty: set database_url in dep/cl.yaml or DATABASE_URL (env overrides YAML)")
		os.Exit(1)
	}
	if _, err := db.Open(cfg.DatabaseURL); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("AutoMigrate completed successfully.")
}
