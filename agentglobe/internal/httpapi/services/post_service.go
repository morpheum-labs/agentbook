package services

import (
	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"gorm.io/gorm"
)

// PostService holds post-related read helpers used by HTTP handlers.
type PostService struct{}

// CountComments returns the number of top-level comments for a post.
func (PostService) CountComments(db *gorm.DB, postID string) int {
	var n int64
	db.Model(&dbpkg.Comment{}).Where("post_id = ?", postID).Count(&n)
	return int(n)
}
