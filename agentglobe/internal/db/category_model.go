package db

import (
	"strings"
	"time"
)

// CategoryUncategorized is the fallback id when a legacy row had an empty label.
const CategoryUncategorized = "uncategorized"

// Category is a managed taxonomy row (topic class / service kind). Primary key is a stable
// string (often equal to the human label, e.g. SPORT/NBA, news) used as FK in floor and capability tables.
type Category struct {
	ID          string    `gorm:"primaryKey;type:text"`
	DisplayName string    `gorm:"not null;type:text"`
	SortOrder   int       `gorm:"not null;default:0;index"`
	IsActive    bool      `gorm:"not null;default:true;index"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (Category) TableName() string { return "categories" }

// FloorCategoryLabel returns the display string for API maps (join or category_id only).
func (q *FloorQuestion) FloorCategoryLabel() string {
	if q == nil {
		return ""
	}
	if q.Category.ID != "" && q.Category.DisplayName != "" {
		return q.Category.DisplayName
	}
	if strings.TrimSpace(q.CategoryID) != "" {
		return q.CategoryID
	}
	return ""
}

// FloorProposalCategoryLabel returns the display string for a topic proposal.
func (p *FloorTopicProposal) FloorProposalCategoryLabel() string {
	if p == nil {
		return ""
	}
	if p.Category.ID != "" && p.Category.DisplayName != "" {
		return p.Category.DisplayName
	}
	if strings.TrimSpace(p.CategoryID) != "" {
		return p.CategoryID
	}
	return ""
}

// CapabilityCategoryLabel returns the optional category for JSON (empty if unset).
func (c *CapabilityService) CapabilityCategoryLabel() string {
	if c == nil {
		return ""
	}
	if c.Category != nil && c.CategoryID != nil && c.Category.ID != "" && c.Category.DisplayName != "" {
		return c.Category.DisplayName
	}
	if c.CategoryID != nil {
		return strings.TrimSpace(*c.CategoryID)
	}
	return ""
}
