package db

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Capability run states reported via registration or heartbeat.
const (
	CapabilityServiceStatusActive   = "active"
	CapabilityServiceStatusDegraded = "degraded"
	CapabilityServiceStatusInactive = "inactive"
)

// DefaultHeartbeatGrace is the window after last_seen in which a service is considered "healthy" for [CapabilityService.IsHealthy].
const DefaultHeartbeatGrace = 5 * time.Minute

// CapabilityService stores a registered API-first service (e.g. newsapi, worldmon) for agent discovery.
type CapabilityService struct {
	ID              string `gorm:"primaryKey;type:text"`
	Name            string `gorm:"not null;type:text;uniqueIndex:ux_capability_service_name_base_url"`
	Version         string `gorm:"type:text;not null"`
	BaseURL         string `gorm:"column:base_url;type:text;not null;uniqueIndex:ux_capability_service_name_base_url"`
	Description     string `gorm:"type:text"`
	CategoryID      *string   `gorm:"column:category_id;type:text;index:idx_cap_svc_category"`
	Category         *Category `gorm:"foreignKey:CategoryID;references:ID"`
	TagsJSON        string `gorm:"column:tags;type:json;not null;default:'[]'"`
	DomainsJSON     string `gorm:"column:domains;type:json;not null;default:'[]'"`
	MetadataJSON    string `gorm:"column:metadata;type:json;not null;default:'{}'"`
	OpenapiURL      string `gorm:"column:openapi_url;type:text"`
	OpenapiSpecJSON string `gorm:"column:openapi_spec;type:text;not null;default:''"`
	Status          string `gorm:"type:text;not null;default:'active';index:idx_cap_svc_status"`
	LastSeen        *time.Time `gorm:"column:last_seen;index:idx_cap_svc_last_seen"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
}

func (CapabilityService) TableName() string { return "capability_services" }

func (c *CapabilityService) BeforeCreate(tx *gorm.DB) error {
	_ = tx
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	if strings.TrimSpace(c.Status) == "" {
		c.Status = CapabilityServiceStatusActive
	} else {
		c.Status = strings.ToLower(strings.TrimSpace(c.Status))
	}
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

// IsHealthy is true when status is not inactive, last_seen is set, and last_seen is within grace of now.
func (c *CapabilityService) IsHealthy(grace time.Duration) bool {
	if c == nil || c.LastSeen == nil {
		return false
	}
	if strings.ToLower(c.Status) == CapabilityServiceStatusInactive {
		return false
	}
	if grace <= 0 {
		grace = DefaultHeartbeatGrace
	}
	return time.Since(*c.LastSeen) < grace
}

// KnownCapabilityServiceStatus returns true for allowed status values.
func KnownCapabilityServiceStatus(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", CapabilityServiceStatusActive, CapabilityServiceStatusDegraded, CapabilityServiceStatusInactive:
		return true
	default:
		return false
	}
}
