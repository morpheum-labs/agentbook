// Schema export: writes DDL for the database selected by config (Postgres via pg_dump, SQLite via sqlite_master).
//
// Usage (from agentglobe/):
//
//	CONFIG_PATH=../dep/cf.yaml go run ./cmd/schemaexport
//	go run ./cmd/schemaexport -out ../../spec/agentglobe_schema.sql
//
// Requires pg_dump on PATH when using database_url (Postgres).
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/morpheumlabs/agentbook/agentglobe/internal/config"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"gorm.io/gorm"
)

func main() {
	out := flag.String("out", "", "output .sql path (default: <repo>/spec/agentglobe_schema.sql)")
	cfgFlag := flag.String("config", "", "config YAML (default: CONFIG_PATH or DefaultConfigPath)")
	flag.Parse()

	cfgPath := strings.TrimSpace(*cfgFlag)
	if cfgPath == "" {
		cfgPath = strings.TrimSpace(os.Getenv("CONFIG_PATH"))
		if cfgPath == "" {
			cfgPath = config.DefaultConfigPath()
		}
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	gdb, err := db.Open(cfg)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}

	outPath := strings.TrimSpace(*out)
	if outPath == "" {
		outPath = filepath.Join(findRepoRoot(), "spec", "agentglobe_schema.sql")
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		log.Fatalf("mkdir: %v", err)
	}

	var exportErr error
	if isPostgres(cfg) {
		exportErr = exportPostgres(cfg.DatabaseURL, outPath)
	} else {
		exportErr = exportSQLite(gdb, outPath)
	}
	if exportErr != nil {
		log.Fatal(exportErr)
	}
	log.Printf("wrote %s", outPath)
}

func isPostgres(cfg *config.Config) bool {
	u := strings.TrimSpace(cfg.DatabaseURL)
	return strings.HasPrefix(u, "postgres://") || strings.HasPrefix(u, "postgresql://")
}

func exportPostgres(dsn, outPath string) error {
	hdr := fmt.Sprintf("-- agentglobe schema dump (pg_dump --schema-only)\n-- generated: %s\n\n", time.Now().UTC().Format(time.RFC3339))
	run := func(cmd *exec.Cmd) error {
		f, err := os.Create(outPath)
		if err != nil {
			return err
		}
		if _, err := f.WriteString(hdr); err != nil {
			_ = f.Close()
			return err
		}
		cmd.Stdout = f
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		_ = f.Close()
		if err != nil {
			_ = os.Remove(outPath)
		}
		return err
	}

	var pgDumpErr error
	if _, err := exec.LookPath("pg_dump"); err == nil {
		pgDumpErr = run(exec.Command("pg_dump", dsn, "--schema-only", "--no-owner", "--no-acl"))
		if pgDumpErr == nil {
			return nil
		}
	}
	if _, err := exec.LookPath("docker"); err == nil {
		derr := run(exec.Command("docker", "run", "--rm", "postgres:17-alpine",
			"pg_dump", dsn, "--schema-only", "--no-owner", "--no-acl"))
		if derr == nil {
			return nil
		}
		if pgDumpErr != nil {
			return fmt.Errorf("pg_dump: %w; docker pg_dump: %v", pgDumpErr, derr)
		}
		return fmt.Errorf("docker pg_dump: %w", derr)
	}
	if pgDumpErr != nil {
		return fmt.Errorf("pg_dump: %w (install postgresql@17 client or add docker for remote PG 17+)", pgDumpErr)
	}
	return fmt.Errorf("need pg_dump on PATH or docker for postgres schema export")
}

func exportSQLite(gdb *gorm.DB, outPath string) error {
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintf(f, "-- agentglobe schema dump (sqlite_master)\n-- generated: %s\n\n", time.Now().UTC().Format(time.RFC3339))

	rows, err := gdb.Raw(`
SELECT sql || ';'
FROM sqlite_master
WHERE sql IS NOT NULL
  AND name NOT LIKE 'sqlite%'
ORDER BY CASE type
  WHEN 'table' THEN 1
  WHEN 'view' THEN 2
  WHEN 'index' THEN 3
  WHEN 'trigger' THEN 4
  ELSE 5
END, tbl_name, name`).Rows()
	if err != nil {
		_ = os.Remove(outPath)
		return fmt.Errorf("sqlite_master: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return err
		}
		if _, err := f.WriteString(s + "\n\n"); err != nil {
			return err
		}
	}
	return rows.Err()
}

func findRepoRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return dir
		}
		dir = parent
	}
}
