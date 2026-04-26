package db

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector
	postgresMode := false
	dbURL := strings.TrimSpace(cfg.DatabaseURL)
	if dbURL == "" {
		dbURL = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}
	if dbURL != "" {
		if strings.HasPrefix(dbURL, "postgres://") || strings.HasPrefix(dbURL, "postgresql://") {
			if ms, ok := envPositiveInt("PG_STATEMENT_TIMEOUT_MS"); ok {
				var err error
				dbURL, err = mergePostgresStatementTimeout(dbURL, ms)
				if err != nil {
					return nil, fmt.Errorf("postgres database_url: %w", err)
				}
			}
			dialector = postgres.Open(dbURL)
			postgresMode = true
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
	if postgresMode {
		if err := configurePostgresPool(gdb); err != nil {
			return nil, err
		}
	}
	if err := gdb.AutoMigrate(
		&Category{},
		&Agent{},
		&Project{},
		&ProjectMember{},
		&Post{},
		&Comment{},
		&Webhook{},
		&GitHubWebhook{},
		&Notification{},
		&Attachment{},
		&FloorQuestion{},
		&FloorExternalSignal{},
		&FloorPosition{},
		&FloorAgentTopicStat{},
		&FloorAgentInferenceProfile{},
		&FloorDigestEntry{},
		&FloorQuestionProbabilityPoint{},
		&FloorPositionChallenge{},
		&FloorResearchArticle{},
		&FloorTopicProposal{},
		&FloorBroadcast{},
		&FloorIndexPageMeta{},
		&FloorIndexEntry{},
		&DebateThread{},
		&DebatePost{},
		&DebatePostReport{},
		&AgentSanction{},
		&CapabilityService{},
		&MCPMemory{},
	); err != nil {
		return nil, err
	}
	if err := MigrateCategoryReferences(gdb); err != nil {
		return nil, err
	}
	if postgresMode {
		if err := createPostgresCapabilityServiceIndexes(gdb); err != nil {
			return nil, err
		}
	}
	return gdb, nil
}

func envPositiveInt(key string) (int, bool) {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

// mergePostgresStatementTimeout appends libpq `options=-c statement_timeout=...` (milliseconds) to the URL.
func mergePostgresStatementTimeout(dsn string, timeoutMS int) (string, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", err
	}
	opt := fmt.Sprintf("-c statement_timeout=%d", timeoutMS)
	q := u.Query()
	if prev := q.Get("options"); prev != "" {
		q.Set("options", prev+" "+opt)
	} else {
		q.Set("options", opt)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// configurePostgresPool sets *sql.DB pool limits for server-side Postgres (managed DB, PgBouncer, etc.).
// Override with PG_MAX_OPEN_CONNS, PG_MAX_IDLE_CONNS, PG_CONN_MAX_LIFETIME, PG_CONN_MAX_IDLE_TIME.
func configurePostgresPool(gdb *gorm.DB) error {
	sqlDB, err := gdb.DB()
	if err != nil {
		return err
	}
	maxOpen := 64
	if s := strings.TrimSpace(os.Getenv("PG_MAX_OPEN_CONNS")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			maxOpen = n
		}
	}
	maxIdle := min(16, maxOpen)
	if s := strings.TrimSpace(os.Getenv("PG_MAX_IDLE_CONNS")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n >= 0 {
			maxIdle = n
			if maxIdle > maxOpen {
				maxIdle = maxOpen
			}
		}
	}
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	lifetime := 30 * time.Minute
	if s := strings.TrimSpace(os.Getenv("PG_CONN_MAX_LIFETIME")); s != "" {
		if d, err := time.ParseDuration(s); err == nil {
			lifetime = d
		}
	}
	if lifetime > 0 {
		sqlDB.SetConnMaxLifetime(lifetime)
	} else {
		sqlDB.SetConnMaxLifetime(0)
	}
	idleTime := 5 * time.Minute
	if s := strings.TrimSpace(os.Getenv("PG_CONN_MAX_IDLE_TIME")); s != "" {
		if d, err := time.ParseDuration(s); err == nil {
			idleTime = d
		}
	}
	if idleTime > 0 {
		sqlDB.SetConnMaxIdleTime(idleTime)
	} else {
		sqlDB.SetConnMaxIdleTime(0)
	}
	return nil
}
