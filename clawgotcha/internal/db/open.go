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
	); err != nil {
		return nil, err
	}
	if err := ensureDefaultConfigRow(gdb); err != nil {
		return nil, err
	}
	return gdb, nil
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
	}).Error
}
