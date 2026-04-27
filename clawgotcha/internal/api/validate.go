package api

import (
	"fmt"
	"strings"

	"github.com/morpheumlabs/agentbook/clawgotcha/internal/db"
	"gorm.io/gorm"
)

func validateAutonomy(s string) error {
	switch strings.TrimSpace(s) {
	case db.AutonomyReadOnly, db.AutonomySupervised, db.AutonomyFull:
		return nil
	default:
		return fmt.Errorf("autonomy_level must be one of: ReadOnly, Supervised, Full")
	}
}

// agentExists returns nil if a swarm agent with the given name exists, otherwise
// gorm.ErrRecordNotFound.
func agentExists(gdb *gorm.DB, name string) error {
	var n int64
	err := gdb.Model(&db.SwarmAgent{}).Where("name = ?", name).Count(&n).Error
	if err != nil {
		return err
	}
	if n == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
