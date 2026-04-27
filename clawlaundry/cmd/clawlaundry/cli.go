package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/morpheumlabs/agentbook/clawlaundry/internal/api"
	"github.com/morpheumlabs/agentbook/clawlaundry/internal/config"
	"github.com/morpheumlabs/agentbook/clawlaundry/internal/db"
	"github.com/morpheumlabs/agentbook/clawlaundry/internal/prompt"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var (
	configPath string
	workspaceW string
)

func newRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "clawlaundry",
		Short: "HTTP API for MiroClaw/ZeroClaw-style swarm agent metadata, plus prompt workspace helpers",
		Long:  "PostgreSQL + GORM backend. With no subcommand, starts the HTTP server (config via -c/--config or CONFIG_PATH).",
		RunE:  runServer,
	}
	root.PersistentFlags().StringVarP(&configPath, "config", "c", "", "YAML config (e.g. dep/cl.yaml); default: CONFIG_PATH or search dep/cl.yaml, ../dep/cl.yaml")
	root.AddCommand(newPromptCmd())
	return root
}

func newPromptCmd() *cobra.Command {
	compose := &cobra.Command{
		Use:   "compose AGENT_NAME",
		Short: "Read IDENTITY.md + SOUL.md + USER.md in --workspace, update the agent's system_prompt in the DB",
		Args:  cobra.ExactArgs(1),
		RunE:  runPromptCompose,
	}
	decompose := &cobra.Command{
		Use:   "decompose AGENT_NAME",
		Short: "Read system_prompt for the agent from the DB, write the three MiroClaw files under --workspace",
		Args:  cobra.ExactArgs(1),
		RunE:  runPromptDecompose,
	}
	compose.Flags().StringVarP(&workspaceW, "workspace", "w", "", "MiroClaw workspace directory containing IDENTITY.md, SOUL.md, USER.md")
	decompose.Flags().StringVarP(&workspaceW, "workspace", "w", "", "Target directory for IDENTITY.md, SOUL.md, USER.md (created if needed)")
	_ = compose.MarkFlagRequired("workspace")
	_ = decompose.MarkFlagRequired("workspace")
	cmd := &cobra.Command{
		Use:   "prompt",
		Short: "Modular MiroClaw prompt files (IDENTITY, SOUL, USER) and DB system_prompt",
	}
	cmd.AddCommand(compose, decompose)
	return cmd
}

func resolveConfigPath() string {
	if p := strings.TrimSpace(configPath); p != "" {
		return p
	}
	if p := strings.TrimSpace(os.Getenv("CONFIG_PATH")); p != "" {
		return p
	}
	return config.DefaultConfigPath()
}

func openDB() (*gorm.DB, error) {
	p := resolveConfigPath()
	cfg, err := config.Load(p)
	if err != nil {
		return nil, err
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("database is required: set database_url in YAML (e.g. dep/cl.yaml) or DATABASE_URL")
	}
	return db.Open(cfg.DatabaseURL)
}

func runServer(_ *cobra.Command, _ []string) error {
	p := resolveConfigPath()
	cfg, err := config.Load(p)
	if err != nil {
		return err
	}
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("database is required: set database_url in YAML (e.g. dep/cl.yaml) or DATABASE_URL")
	}
	g, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		return err
	}
	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: api.NewRouter(g)}
	slog.Info("clawlaundry listening", "addr", cfg.HTTPAddr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

func runPromptCompose(_ *cobra.Command, args []string) error {
	g, err := openDB()
	if err != nil {
		return err
	}
	agentName := strings.TrimSpace(args[0])
	if agentName == "" {
		return fmt.Errorf("AGENT_NAME required")
	}
	combined, err := prompt.Compose(strings.TrimSpace(workspaceW))
	if err != nil {
		return err
	}
	tx := g.Model(&db.SwarmAgent{}).Where("name = ?", agentName).Update("system_prompt", combined)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return fmt.Errorf("no agent with name %q", agentName)
	}
	_, _ = fmt.Fprintf(os.Stdout, "updated system_prompt for %q from %s (IDENTITY, SOUL, USER)\n", agentName, strings.TrimSpace(workspaceW))
	return nil
}

func runPromptDecompose(_ *cobra.Command, args []string) error {
	g, err := openDB()
	if err != nil {
		return err
	}
	agentName := strings.TrimSpace(args[0])
	if agentName == "" {
		return fmt.Errorf("AGENT_NAME required")
	}
	var a db.SwarmAgent
	if err := g.Where("name = ?", agentName).First(&a).Error; err != nil {
		return err
	}
	ws := strings.TrimSpace(workspaceW)
	if err := prompt.Decompose(a.SystemPrompt, ws); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(os.Stdout, "wrote %s, %s, %s under %s from agent %q system_prompt\n",
		prompt.FilenameIdentity, prompt.FilenameSoul, prompt.FilenameUser, ws, agentName)
	return nil
}
