package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const HeartbeatStaleSeconds = 90

// MarkStaleRuntimeInstancesOffline sets status=offline for instances whose last heartbeat
// is older than HeartbeatStaleSeconds, and disables their webhook subscriptions.
func MarkStaleRuntimeInstancesOffline(gdb *gorm.DB) error {
	cutoff := time.Now().UTC().Add(-HeartbeatStaleSeconds * time.Second)
	return gdb.Transaction(func(tx *gorm.DB) error {
		var ids []uuid.UUID
		err := tx.Model(&SwarmRuntimeInstance{}).
			Where("last_heartbeat_at IS NOT NULL AND last_heartbeat_at < ?", cutoff).
			Where("status IN ?", []string{RuntimeStatusOnline, RuntimeStatusDegraded, RuntimeStatusUnknown}).
			Pluck("id", &ids).Error
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		if err := tx.Model(&SwarmRuntimeInstance{}).Where("id IN ?", ids).Updates(map[string]interface{}{
			"status": RuntimeStatusOffline,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&SwarmWebhookSubscription{}).Where("runtime_instance_id IN ?", ids).Update("enabled", false).Error
	})
}
