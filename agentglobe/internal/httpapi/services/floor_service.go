package services

import (
	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"gorm.io/gorm"
)

// FloorService holds shared read helpers for HTTP handlers (AgentFloor–adjacent and forum surfaces).
type FloorService struct{}

// CountComments returns the number of top-level comments for a post.
func (FloorService) CountComments(db *gorm.DB, postID string) int {
	var n int64
	db.Model(&dbpkg.Comment{}).Where("post_id = ?", postID).Count(&n)
	return int(n)
}
