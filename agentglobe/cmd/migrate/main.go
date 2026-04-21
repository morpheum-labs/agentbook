// Command migrate applies ordered SQL migrations from a directory to PostgreSQL.
// It records successful runs in public.schema_migrations (name = file basename).
// Safe to re-run: each file runs at most once.
//
// Database URL resolution (same as agentglobe server):
//   1) DATABASE_URL env (if set) overrides YAML
//   2) database_url from config YAML (-c / -config path, else CONFIG_PATH env, else config.DefaultConfigPath)
//
// Typical use from repo root (see Makefile migrate target):
//
//	make migrate
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/morpheumlabs/agentbook/agentglobe/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const ensureMigrationsTable = `
CREATE TABLE IF NOT EXISTS public.schema_migrations (
	name text PRIMARY KEY,
	applied_at timestamptz NOT NULL DEFAULT now()
);
`

func main() {
	log.SetFlags(0)
	log.SetPrefix("migrate: ")

	defaultDir := strings.TrimSpace(os.Getenv("MIGRATIONS_DIR"))
	if defaultDir == "" {
		defaultDir = "../spec/migrations"
	}

	var cfgPathFromFlags string
	flag.StringVar(&cfgPathFromFlags, "c", "", "config YAML path (e.g. ../dep/cf.yaml); same as -config")
	flag.StringVar(&cfgPathFromFlags, "config", "", "same as -c")
	var dirFromFlags string
	flag.StringVar(&dirFromFlags, "d", defaultDir, "migrations directory containing *.sql files (lexical order by filename)")
	flag.StringVar(&dirFromFlags, "dir", defaultDir, "same as -d")
	dry := flag.Bool("dry-run", false, "print pending migrations and exit without applying")
	flag.Parse()

	migrationsDir := strings.TrimSpace(dirFromFlags)
	if migrationsDir == "" {
		migrationsDir = defaultDir
	}

	cfgPath := strings.TrimSpace(cfgPathFromFlags)
	if cfgPath == "" {
		cfgPath = strings.TrimSpace(os.Getenv("CONFIG_PATH"))
	}
	if cfgPath == "" {
		cfgPath = config.DefaultConfigPath()
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	dsn := strings.TrimSpace(cfg.DatabaseURL)
	if dsn == "" {
		if cfgPath != "" {
			log.Fatalf("database_url is empty in config %q (set database_url for Postgres, or set DATABASE_URL)", cfgPath)
		}
		log.Fatal("database_url is empty: use -c path/to.yaml (or CONFIG_PATH / DATABASE_URL)")
	}
	if !strings.HasPrefix(dsn, "postgres://") && !strings.HasPrefix(dsn, "postgresql://") {
		log.Fatal("database_url must be postgres:// or postgresql:// (SQL migrations are Postgres-only)")
	}
	if cfgPath != "" {
		if abs, err := filepath.Abs(cfgPath); err == nil {
			fmt.Printf("migrate: config=%s\n", abs)
		} else {
			fmt.Printf("migrate: config=%s\n", cfgPath)
		}
	} else {
		fmt.Println("migrate: config=(none; DATABASE_URL from environment only)")
	}

	absDir, err := filepath.Abs(migrationsDir)
	if err != nil {
		log.Fatalf("migrations dir: %v", err)
	}
	entries, err := os.ReadDir(absDir)
	if err != nil {
		log.Fatalf("read migrations dir %q: %v", absDir, err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".sql") {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)
	if len(files) == 0 {
		log.Fatalf("no .sql files in %q", absDir)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	db.SetConnMaxLifetime(time.Minute * 5)
	db.SetMaxOpenConns(2)

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping: %v", err)
	}

	if _, err := db.ExecContext(ctx, ensureMigrationsTable); err != nil {
		log.Fatalf("schema_migrations table: %v", err)
	}

	applied, err := loadApplied(ctx, db)
	if err != nil {
		log.Fatalf("list applied: %v", err)
	}

	var pending []string
	for _, name := range files {
		if !applied[name] {
			pending = append(pending, name)
		}
	}

	if len(pending) == 0 {
		fmt.Println("migrate: nothing pending")
		return
	}

	fmt.Printf("migrate: dir=%s pending=%d\n", absDir, len(pending))
	for _, name := range pending {
		fmt.Println("  -", name)
	}
	if *dry {
		return
	}

	for _, name := range pending {
		path := filepath.Join(absDir, name)
		body, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("read %s: %v", name, err)
		}
		sqlText := strings.TrimSpace(string(body))
		if sqlText == "" {
			log.Fatalf("%s: empty file", name)
		}

		// Migration scripts use their own BEGIN/COMMIT; do not wrap in a driver transaction.
		if _, err := db.ExecContext(ctx, sqlText); err != nil {
			log.Fatalf("%s: exec: %v", name, err)
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO public.schema_migrations (name) VALUES ($1)`, name); err != nil {
			log.Fatalf("%s: record migration (SQL already applied — fix DB or schema_migrations manually): %v", name, err)
		}
		fmt.Printf("migrate: applied %s\n", name)
	}
}

func loadApplied(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT name FROM public.schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out[name] = true
	}
	return out, rows.Err()
}
