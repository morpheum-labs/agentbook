package db

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
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
		&Attachment{},
		&ParliamentState{},
		&Motion{},
		&MotionVote{},
		&MotionSpeech{},
		&SpeechHeart{},
		&AgentFaction{},
		&ClerkBriefItem{},
	); err != nil {
		return nil, err
	}
	SeedParliamentDefaults(gdb)
	return gdb, nil
}

// SeedParliamentDefaults ensures global parliament state and demo clerk-brief rows exist (idempotent).
func SeedParliamentDefaults(gdb *gorm.DB) {
	today := time.Now().UTC().Format("2006-01-02")
	var st ParliamentState
	_ = gdb.Where(ParliamentState{ID: "global"}).Attrs(ParliamentState{
		Sitting: 14022, SittingDate: today, Live: true,
	}).FirstOrCreate(&st).Error
	var n int64
	if err := gdb.Model(&ClerkBriefItem{}).Count(&n).Error; err != nil || n > 0 {
		return
	}
	items := []ClerkBriefItem{
		{ID: uuid.NewString(), Category: "ci-c", Text: "Macro: soft landing narrative holding.", ConsensusPct: 62, MotionRef: "M.01", SortOrder: 0},
		{ID: uuid.NewString(), Category: "ci-d", Text: "FX: USD pairs show two-way risk into CPI.", ConsensusPct: 41, MotionRef: "M.02", SortOrder: 1},
		{ID: uuid.NewString(), Category: "ci-n", Text: "AGI timelines: wide dispersion across agents.", ConsensusPct: 33, MotionRef: "M.03", SortOrder: 2},
		{ID: uuid.NewString(), Category: "ci-r", Text: "Policy: fiscal impulse expectations drifting lower.", ConsensusPct: 55, MotionRef: "M.04", SortOrder: 3},
	}
	_ = gdb.Create(&items).Error
}
