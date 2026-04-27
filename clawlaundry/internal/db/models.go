package db

import (
	"time"

	"github.com/google/uuid"
)

// SwarmConfig is a single-row table mirroring the top-level defaults in agentic_swarm
// (default_provider, default_model). ID is always 1.
type SwarmConfig struct {
	ID              uint   `gorm:"primaryKey"`
	DefaultProvider string `gorm:"not null;type:text;column:default_provider"`
	DefaultModel    string `gorm:"not null;type:text;column:default_model"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (SwarmConfig) TableName() string { return "swarm_config" }

// Autonomy* mirror agentic_swarm examples: ReadOnly, Supervised, Full.
const (
	AutonomyReadOnly   = "ReadOnly"
	AutonomySupervised = "Supervised"
	AutonomyFull       = "Full"
)

// SwarmAgent is one Hand (one [[agents]] block): name, system_prompt, tools, provider, model, etc.
type SwarmAgent struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name            string    `gorm:"uniqueIndex;not null;type:text"`
	SystemPrompt    string    `gorm:"not null;type:text;column:system_prompt"`
	Tools           []string  `gorm:"serializer:json;type:jsonb;not null;default:'[]'"`
	Provider        string    `gorm:"not null;type:text"`
	Model           string    `gorm:"not null;type:text"`
	TimeoutSeconds  int       `gorm:"not null;column:timeout_seconds"`
	AutonomyLevel   string    `gorm:"not null;type:text;column:autonomy_level"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (SwarmAgent) TableName() string { return "swarm_agents" }

// SwarmCronJob is one [[cron_jobs]] block: name, target agent, schedule, prompt, timeout.
type SwarmCronJob struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name           string    `gorm:"uniqueIndex;not null;type:text"`
	AgentName      string    `gorm:"not null;type:text;index;column:agent_name"`
	Schedule       string    `gorm:"not null;type:text"`
	TimeoutSeconds int       `gorm:"not null;column:timeout_seconds"`
	Prompt         string    `gorm:"not null;type:text"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (SwarmCronJob) TableName() string { return "swarm_cron_jobs" }
