package config

import (
	"os"
	"path/filepath"
	"testing"
)

func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"DATABASE_URL", "SQLITE_PATH", "ADMIN_TOKEN", "SERVICE_REGISTRY_TOKEN", "PUBLIC_URL", "HOSTNAME",
		"PORT", "ATTACHMENTS_DIR", "MAX_ATTACHMENT_BYTES", "CORS_ALLOWED_ORIGINS",
	} {
		t.Setenv(k, "")
	}
}

func TestLoad_defaultsWhenNoFile(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("PORT", "")
	t.Setenv("HOSTNAME", "")
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 3456 || cfg.Hostname != "localhost:3456" {
		t.Fatalf("defaults: got port=%d hostname=%q", cfg.Port, cfg.Hostname)
	}
	if cfg.Database != "data/minibook.db" {
		t.Fatalf("default database: got %q", cfg.Database)
	}
}

func TestLoad_yamlOverridesDefaults(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("PORT", "")
	t.Setenv("HOSTNAME", "")
	dir := t.TempDir()
	p := filepath.Join(dir, "cfg.yaml")
	if err := os.WriteFile(p, []byte("port: 4000\nhostname: yaml-host:4000\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 4000 || cfg.Hostname != "yaml-host:4000" {
		t.Fatalf("yaml: got port=%d hostname=%q", cfg.Port, cfg.Hostname)
	}
}

func TestLoad_envOverridesYAML(t *testing.T) {
	clearConfigEnv(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "cfg.yaml")
	if err := os.WriteFile(p, []byte("port: 4000\nadmin_token: from-yaml\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PORT", "5000")
	t.Setenv("ADMIN_TOKEN", "from-env")
	cfg, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 5000 {
		t.Fatalf("PORT: want 5000 got %d", cfg.Port)
	}
	if cfg.AdminToken != "from-env" {
		t.Fatalf("ADMIN_TOKEN: want from-env got %q", cfg.AdminToken)
	}
}

func TestLoad_databaseURLEnvOverridesYAMLDatabase(t *testing.T) {
	clearConfigEnv(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "cfg.yaml")
	if err := os.WriteFile(p, []byte("database_url: \"postgresql://yaml/db\"\ndatabase: \"from-yaml.db\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DATABASE_URL", "postgresql://env/db")
	t.Setenv("SQLITE_PATH", "env-sqlite.db")
	cfg, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DatabaseURL != "postgresql://env/db" {
		t.Fatalf("DATABASE_URL: got %q", cfg.DatabaseURL)
	}
	// SQLITE_PATH must not apply when postgres URL is set
	if cfg.Database != "from-yaml.db" {
		t.Fatalf("database with postgres URL: want from-yaml.db got %q", cfg.Database)
	}
}

func TestLoad_sqlitePathWhenNoDatabaseURL(t *testing.T) {
	clearConfigEnv(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "cfg.yaml")
	if err := os.WriteFile(p, []byte("database: \"from-yaml.db\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DATABASE_URL", "")
	t.Setenv("SQLITE_PATH", "from-env.db")
	cfg, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Database != "from-env.db" {
		t.Fatalf("SQLITE_PATH: want from-env.db got %q", cfg.Database)
	}
}

func TestLoad_serviceRegistryTokenFromEnv(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("SERVICE_REGISTRY_TOKEN", "reg-from-env")
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ServiceRegistryToken != "reg-from-env" {
		t.Fatalf("got %q", cfg.ServiceRegistryToken)
	}
}
