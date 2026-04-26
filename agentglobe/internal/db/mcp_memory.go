package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MCPMemory stores key-value text blobs for MCP clients (e.g. news_curator) separate from miroclaw's local memory.
type MCPMemory struct {
	ID          string     `gorm:"primaryKey;type:text"`
	AgentID     string     `gorm:"column:agent_id;index;not null;type:text;uniqueIndex:ux_mcp_mem_agent_ns_key"`
	Namespace   string     `gorm:"type:text;not null;default:'';uniqueIndex:ux_mcp_mem_agent_ns_key"`
	Key         string     `gorm:"column:mcp_key;type:text;not null;uniqueIndex:ux_mcp_mem_agent_ns_key"`
	Content     string     `gorm:"type:text"`
	TagsJSON    string     `gorm:"column:tags;type:text;not null;default:'[]'"`
	ExpiresAt   *time.Time `gorm:"column:expires_at;index:idx_mcp_mem_expires"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
	Agent       Agent      `gorm:"foreignKey:AgentID;references:ID"`
}

// TableName is gorm table name.
func (MCPMemory) TableName() string { return "mcp_memories" }

// BeforeCreate sets id if empty.
func (m *MCPMemory) BeforeCreate(tx *gorm.DB) error {
	_ = tx
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	if m.TagsJSON == "" {
		m.TagsJSON = "[]"
	}
	return nil
}

// SetTags encodes tag slice into TagsJSON.
func (m *MCPMemory) SetTags(tags []string) {
	m.TagsJSON = encodeStringSlice(tags)
}

// TagSlice decodes tags.
func (m *MCPMemory) TagSlice() []string { return decodeStringSlice(m.TagsJSON) }
