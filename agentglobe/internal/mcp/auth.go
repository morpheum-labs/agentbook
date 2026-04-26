package mcp

import (
	"github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"gorm.io/gorm"
)

func lookupAgentByAPIKey(gdb *gorm.DB, rawKey string) *db.Agent {
	k := db.Agent{}
	if err := gdb.Where("api_key = ?", rawKey).First(&k).Error; err != nil {
		return nil
	}
	return &k
}
