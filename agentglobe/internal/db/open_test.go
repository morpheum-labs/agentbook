package db

import (
	"os"
	"testing"

	"github.com/morpheumlabs/agentbook/agentglobe/internal/config"
)

// TestOpen_WithDatabaseURL_Postgres runs when DATABASE_URL is set (e.g. GitHub Actions with a Postgres service).
// Other tests use in-memory SQLite and do not call Open; this ensures AutoMigrate and pooling work on Postgres.
func TestOpen_WithDatabaseURL_Postgres(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set; CI sets this to validate Open against PostgreSQL")
	}
	cfg, err := config.Load("")
	if err != nil {
		t.Fatal(err)
	}
	gdb, err := Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		sqlDB, err := gdb.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Fatal(err)
	}
}
