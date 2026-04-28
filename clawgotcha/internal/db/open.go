package db

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open connects to PostgreSQL and runs AutoMigrate for swarm tables.
func Open(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	if err := gdb.AutoMigrate(
		&SwarmConfig{},
		&SwarmAgent{},
		&SwarmCronJob{},
		&SwarmRuntimeInstance{},
		&SwarmWebhookSubscription{},
	); err != nil {
		return nil, err
	}
	if err := ensurePartialUniqueNameIndexes(gdb); err != nil {
		return nil, err
	}
	if err := ensureDefaultConfigRow(gdb); err != nil {
		return nil, err
	}
	return gdb, nil
}

// ensurePartialUniqueNameIndexes enforces unique agent/cron names among non-deleted rows only.
func ensurePartialUniqueNameIndexes(gdb *gorm.DB) error {
	stmts := []string{
		`DROP INDEX IF EXISTS idx_swarm_agents_name`,
		`ALTER TABLE swarm_agents DROP CONSTRAINT IF EXISTS swarm_agents_name_key`,
		`CREATE UNIQUE INDEX IF NOT EXISTS swarm_agents_name_alive_idx ON swarm_agents (name) WHERE deleted_at IS NULL`,
		`DROP INDEX IF EXISTS idx_swarm_cron_jobs_name`,
		`ALTER TABLE swarm_cron_jobs DROP CONSTRAINT IF EXISTS swarm_cron_jobs_name_key`,
		`CREATE UNIQUE INDEX IF NOT EXISTS swarm_cron_jobs_name_alive_idx ON swarm_cron_jobs (name) WHERE deleted_at IS NULL`,
	}
	for _, q := range stmts {
		if err := gdb.Exec(q).Error; err != nil {
			return err
		}
	}
	return nil
}

func ensureDefaultConfigRow(gdb *gorm.DB) error {
	var n int64
	if err := gdb.Model(&SwarmConfig{}).Where("id = ?", 1).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return gdb.Create(&SwarmConfig{
		ID:              1,
		DefaultProvider: "openai",
		DefaultModel:    "",
		CurrentRevision: 1,
	}).Error
}
