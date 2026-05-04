package db

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SwarmConfig is a single-row table mirroring the top-level defaults in agentic_swarm
// (default_provider, default_model). ID is always 1.
type SwarmConfig struct {
	ID              uint   `gorm:"primaryKey"`
	DefaultProvider string `gorm:"not null;type:text;column:default_provider"`
	DefaultModel    string `gorm:"not null;type:text;column:default_model"`
	CurrentRevision int64  `gorm:"not null;default:1;column:current_revision"`
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

// Runtime status for Miroclaw instances.
const (
	RuntimeStatusOnline   = "online"
	RuntimeStatusOffline  = "offline"
	RuntimeStatusDegraded = "degraded"
	RuntimeStatusUnknown  = "unknown"
)

// SwarmAgent is one Hand (one [[agents]] block): name, system_prompt, tools, provider, model, etc.
type SwarmAgent struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name            string     `gorm:"not null;type:text;index"`
	SystemPrompt    string     `gorm:"not null;type:text;column:system_prompt"`
	Tools           []string   `gorm:"serializer:json;type:jsonb;not null;default:'[]'"`
	Provider        string     `gorm:"not null;type:text"`
	Model           string     `gorm:"not null;type:text"`
	TimeoutSeconds  int        `gorm:"not null;column:timeout_seconds"`
	AutonomyLevel   string     `gorm:"not null;type:text;column:autonomy_level"`
	CurrentRevision int64      `gorm:"not null;default:1;column:current_revision;index"`
	LastChangedAt   *time.Time `gorm:"column:last_changed_at"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SwarmAgent) TableName() string { return "swarm_agents" }

// SwarmCronJob is one [[cron_jobs]] block: name, target agent, schedule, prompt, timeout.
type SwarmCronJob struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name            string     `gorm:"not null;type:text;index"`
	AgentName       string     `gorm:"not null;type:text;index;column:agent_name"`
	Schedule        string     `gorm:"not null;type:text"`
	TimeoutSeconds  int        `gorm:"not null;column:timeout_seconds"`
	Prompt          string     `gorm:"not null;type:text"`
	Active          bool       `gorm:"not null;default:true;column:active"`
	CurrentRevision int64      `gorm:"not null;default:1;column:current_revision;index"`
	LastChangedAt   *time.Time `gorm:"column:last_changed_at"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SwarmCronJob) TableName() string { return "swarm_cron_jobs" }

// SwarmRuntimeInstance is the registry row for a Miroclaw runtime.
type SwarmRuntimeInstance struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	InstanceName    string          `gorm:"uniqueIndex;not null;type:text;column:instance_name"`
	InstanceType    string          `gorm:"not null;default:miroclaw;type:text;column:instance_type"`
	Version         string          `gorm:"not null;type:text"`
	Hostname        string          `gorm:"not null;type:text"`
	PublicURL       *string         `gorm:"type:text;column:public_url"`
	CallbackURL     string          `gorm:"not null;type:text;column:callback_url"`
	Capabilities    []string        `gorm:"serializer:json;type:jsonb;not null;default:'[]'"`
	LastHeartbeatAt *time.Time      `gorm:"column:last_heartbeat_at;index"`
	Status          string          `gorm:"not null;default:unknown;type:text;index"`
	StartedAt       time.Time       `gorm:"not null;column:started_at"`
	Metadata        json.RawMessage `gorm:"type:jsonb"`
	// ApiSecretHash is SHA256(instance_api_secret) as 64-char lowercase hex (constant-time verify).
	ApiSecretHash     string     `gorm:"type:text;column:api_secret_hash" json:"-"`
	ApiSecretPrefix   string     `gorm:"type:text;column:api_secret_prefix" json:"api_secret_prefix,omitempty"`
	ApiSecretIssuedAt *time.Time `gorm:"column:api_secret_issued_at" json:"api_secret_issued_at,omitempty"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (SwarmRuntimeInstance) TableName() string { return "swarm_runtime_instances" }

// SwarmWebhookSubscription is push subscription state for a runtime (callback delivery).
type SwarmWebhookSubscription struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	RuntimeInstanceID uuid.UUID `gorm:"type:uuid;not null;index;column:runtime_instance_id"`
	EventTypes        []string  `gorm:"serializer:json;type:jsonb;not null;default:'[]';column:event_types"`
	Secret            string    `gorm:"not null;type:text"`
	Enabled           bool      `gorm:"not null;default:true"`
	CreatedAt         time.Time
	UpdatedAt         time.Time

	Runtime *SwarmRuntimeInstance `gorm:"foreignKey:RuntimeInstanceID"`
}

func (SwarmWebhookSubscription) TableName() string { return "swarm_webhook_subscriptions" }

// CredentialBinding groups one integration's metadata for a Hand (SwarmAgent).
// Secret bytes live in CredentialSecretVersion rows (encrypted).
type CredentialBinding struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SwarmAgentID   uuid.UUID       `gorm:"type:uuid;not null;index;column:swarm_agent_id"`
	ProviderSlug   string          `gorm:"not null;type:text;column:provider_slug"`
	Label          string          `gorm:"not null;type:text"`
	McpServerName  *string         `gorm:"type:text;column:mcp_server_name"`
	Metadata       json.RawMessage `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Agent *SwarmAgent `gorm:"foreignKey:SwarmAgentID;references:ID"`
}

func (CredentialBinding) TableName() string { return "credential_bindings" }

// CredentialSecretVersion is one encrypted material snapshot for a binding (rotation = new row, higher version).
type CredentialSecretVersion struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	BindingID     uuid.UUID `gorm:"type:uuid;not null;index;column:binding_id;uniqueIndex:ux_cred_secret_binding_version"`
	Version       int       `gorm:"not null;uniqueIndex:ux_cred_secret_binding_version"`
	MaterialKind  string    `gorm:"not null;type:text;column:material_kind"`
	Ciphertext    []byte    `gorm:"not null;type:bytea;column:ciphertext"`
	Nonce         []byte    `gorm:"not null;type:bytea;column:nonce"`
	KekID         string    `gorm:"not null;type:text;default:'env:v1';column:kek_id"`
	ExpiresAt     *time.Time `gorm:"column:expires_at"`
	CreatedAt     time.Time `gorm:"not null;column:created_at"`

	Binding *CredentialBinding `gorm:"foreignKey:BindingID;references:ID"`
}

func (CredentialSecretVersion) TableName() string { return "credential_secret_versions" }
