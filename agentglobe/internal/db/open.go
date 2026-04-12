package db

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/glebarez/sqlite"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector
	dbURL := strings.TrimSpace(cfg.DatabaseURL)
	if dbURL == "" {
		dbURL = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}
	if dbURL != "" {
		if strings.HasPrefix(dbURL, "postgres://") || strings.HasPrefix(dbURL, "postgresql://") {
			dialector = postgres.Open(dbURL)
		} else {
			return nil, fmt.Errorf("unsupported database_url scheme (use postgres:// or leave empty for sqlite)")
		}
	} else {
		dbPath := cfg.Database
		if dbPath == "" {
			dbPath = "data/minibook.db"
		}
		if dir := filepath.Dir(dbPath); dir != "." && dir != "" {
			_ = os.MkdirAll(dir, 0o755)
		}
		dialector = sqlite.Open(dbPath)
	}
	gdb, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	if err := gdb.AutoMigrate(
		&Agent{},
		&Project{},
		&ProjectMember{},
		&Post{},
		&Comment{},
		&Webhook{},
		&GitHubWebhook{},
		&Notification{},
	); err != nil {
		return nil, err
	}
	return gdb, nil
}
