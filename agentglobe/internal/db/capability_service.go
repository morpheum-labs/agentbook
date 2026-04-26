package db

import (
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"
)

// CapabilityService stores a registered API-first service (e.g. newapi, worldmon) for agent discovery.
type CapabilityService struct {
	ID              string  `gorm:"primaryKey;type:text"`
	Name            string  `gorm:"not null;type:text;uniqueIndex:ux_capability_service_name_base_url"`
	Version         string  `gorm:"type:text;not null"`
	BaseURL         string  `gorm:"column:base_url;type:text;not null;uniqueIndex:ux_capability_service_name_base_url"`
	Description     string  `gorm:"type:text"`
	Category        string  `gorm:"type:text"`
	TagsJSON        string  `gorm:"column:tags;type:text;not null;default:'[]'"`
	DomainsJSON     string  `gorm:"column:domains;type:text;not null;default:'[]'"`
	MetadataJSON    string  `gorm:"column:metadata;type:text;not null;default:'{}'"`
	OpenapiURL      string  `gorm:"column:openapi_url;type:text"`
	OpenapiSpecJSON string  `gorm:"column:openapi_spec;type:text;not null;default:''"`
	LastSeen        *time.Time `gorm:"column:last_seen;index"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
}

func (CapabilityService) TableName() string { return "capability_services" }

func (c *CapabilityService) BeforeCreate(tx *gorm.DB) error {
	_ = tx
	if strings.TrimSpace(c.MetadataJSON) == "" {
		c.MetadataJSON = "{}"
	}
	if c.TagsJSON == "" {
		c.TagsJSON = "[]"
	}
	if c.DomainsJSON == "" {
		c.DomainsJSON = "[]"
	}
	return nil
}

// TagSlice decodes the tags column.
func (c *CapabilityService) TagSlice() []string {
	if c == nil || c.TagsJSON == "" {
		return nil
	}
	var s []string
	if err := json.Unmarshal([]byte(c.TagsJSON), &s); err != nil {
		return nil
	}
	return s
}

// DomainsFromJSON decodes the domains column.
func (c *CapabilityService) DomainsFromJSON() []string {
	if c == nil || c.DomainsJSON == "" {
		return nil
	}
	var s []string
	if err := json.Unmarshal([]byte(c.DomainsJSON), &s); err != nil {
		return nil
	}
	return s
}

// MetadataMap decodes metadata.
func (c *CapabilityService) MetadataMap() map[string]any {
	if c == nil || strings.TrimSpace(c.MetadataJSON) == "" {
		return map[string]any{}
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(c.MetadataJSON), &m); err != nil || m == nil {
		return map[string]any{}
	}
	return m
}
