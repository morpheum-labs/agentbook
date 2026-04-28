package db

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RevisionSummary is returned in API metadata for sync and runtime registration.
type RevisionSummary struct {
	ConfigRevision      int64 `json:"config_revision"`
	AgentsMaxRevision   int64 `json:"agents_max_revision"`
	CronJobsMaxRevision int64 `json:"cron_jobs_max_revision"`
}

// LoadRevisionSummary aggregates max revisions from config and swarm tables (includes soft-deleted rows).
func LoadRevisionSummary(gdb *gorm.DB) (RevisionSummary, error) {
	var sum RevisionSummary
	var cfg SwarmConfig
	if err := gdb.Unscoped().First(&cfg, 1).Error; err != nil {
		return sum, err
	}
	sum.ConfigRevision = cfg.CurrentRevision

	var agentsMax sql.NullInt64
	if err := gdb.Raw("SELECT COALESCE(MAX(current_revision), 0) FROM swarm_agents").Scan(&agentsMax).Error; err != nil {
		return sum, err
	}
	if agentsMax.Valid {
		sum.AgentsMaxRevision = agentsMax.Int64
	}

	var cronMax sql.NullInt64
	if err := gdb.Raw("SELECT COALESCE(MAX(current_revision), 0) FROM swarm_cron_jobs").Scan(&cronMax).Error; err != nil {
		return sum, err
	}
	if cronMax.Valid {
		sum.CronJobsMaxRevision = cronMax.Int64
	}

	return sum, nil
}

// IncrementAgentRevision bumps current_revision and last_changed_at for one agent row.
func IncrementAgentRevision(tx *gorm.DB, id uuid.UUID) error {
	now := time.Now().UTC()
	return tx.Model(&SwarmAgent{}).Where("id = ?", id).Updates(map[string]interface{}{
		"current_revision": gorm.Expr("current_revision + 1"),
		"last_changed_at":  now,
	}).Error
}

// IncrementCronRevision bumps current_revision and last_changed_at for one cron row.
func IncrementCronRevision(tx *gorm.DB, id uuid.UUID) error {
	now := time.Now().UTC()
	return tx.Model(&SwarmCronJob{}).Where("id = ?", id).Updates(map[string]interface{}{
		"current_revision": gorm.Expr("current_revision + 1"),
		"last_changed_at":  now,
	}).Error
}

// IncrementConfigRevision bumps swarm_config row id=1.
func IncrementConfigRevision(tx *gorm.DB) error {
	now := time.Now().UTC()
	return tx.Model(&SwarmConfig{}).Where("id = ?", 1).Updates(map[string]interface{}{
		"current_revision": gorm.Expr("current_revision + 1"),
		"updated_at":       now,
	}).Error
}

// TouchAgentRevision sets revision fields on a newly created agent (revision 1).
func TouchAgentRevision(a *SwarmAgent) {
	now := time.Now().UTC()
	a.CurrentRevision = 1
	a.LastChangedAt = &now
}

// TouchCronRevision sets revision fields on a newly created cron job (revision 1).
func TouchCronRevision(c *SwarmCronJob) {
	now := time.Now().UTC()
	c.CurrentRevision = 1
	c.LastChangedAt = &now
}
