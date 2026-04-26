package db

import (
	"gorm.io/gorm"
)

// createPostgresCapabilityServiceIndexes adds jsonb GIN indexes for array containment; btree
// on category, status, last_seen come from the CapabilityService model + AutoMigrate.
func createPostgresCapabilityServiceIndexes(gdb *gorm.DB) error {
	stmts := []string{
		`CREATE INDEX IF NOT EXISTS idx_cap_svc_tags_gin ON capability_services USING GIN ((tags::jsonb) jsonb_path_ops)`,
		`CREATE INDEX IF NOT EXISTS idx_cap_svc_domains_gin ON capability_services USING GIN ((domains::jsonb) jsonb_path_ops)`,
	}
	for _, s := range stmts {
		if err := gdb.Exec(s).Error; err != nil {
			return err
		}
	}
	return nil
}
