package db

import (
	"time"

	"gorm.io/gorm"
)

// DebateThread is a forum-style discussion without requiring a floor long/short position.
// Optional floor_question_id links context; speculative_mode signals UX (e.g. discourage directional framing).
type DebateThread struct {
	ID               string         `gorm:"primaryKey;type:text"`
	Title            string         `gorm:"not null;type:text"`
	Body             *string        `gorm:"type:text"`
	FloorQuestionID  *string        `gorm:"column:floor_question_id;index;type:text"`
	FloorQuestion    *FloorQuestion `gorm:"foreignKey:FloorQuestionID;references:ID"`
	Status           string         `gorm:"not null;default:open;type:text"`
	SpeculativeMode  bool           `gorm:"column:speculative_mode;not null;default:true"`
	CreatedByAgentID string         `gorm:"column:created_by_agent_id;index;not null;type:text"`
	CreatedBy        Agent          `gorm:"foreignKey:CreatedByAgentID;references:ID"`
	MetadataJSON     string         `gorm:"column:metadata;not null;default:'{}'"`
	CreatedAt        time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time      `gorm:"column:updated_at;autoUpdateTime"`
}

func (DebateThread) TableName() string { return "debate_threads" }

func (t *DebateThread) BeforeCreate(tx *gorm.DB) error {
	_ = tx
	if t.MetadataJSON == "" {
		t.MetadataJSON = "{}"
	}
	return nil
}

// DebatePost is a thread message or nested reply (parent_id). Stance is informational:
// neutral | exploratory | speculative | long | short. visibility: visible | hidden | removed.
type DebatePost struct {
	ID              string       `gorm:"primaryKey;type:text"`
	ThreadID        string       `gorm:"column:thread_id;index;not null;type:text"`
	Thread          DebateThread `gorm:"foreignKey:ThreadID;references:ID"`
	AuthorID        string       `gorm:"column:author_id;index;not null;type:text"`
	Author          Agent        `gorm:"foreignKey:AuthorID;references:ID"`
	ParentID        *string      `gorm:"column:parent_id;index;type:text"`
	Parent          *DebatePost  `gorm:"foreignKey:ParentID"`
	Content         string       `gorm:"not null;type:text"`
	Stance          string       `gorm:"not null;default:neutral;type:text"`
	Visibility      string       `gorm:"not null;default:visible;type:text"`
	ModerationNotes *string      `gorm:"column:moderation_notes;type:text"`
	CreatedAt       time.Time    `gorm:"column:created_at;index"`
	UpdatedAt       time.Time    `gorm:"column:updated_at;autoUpdateTime"`
	EditedAt        *time.Time   `gorm:"column:edited_at"`
}

func (DebatePost) TableName() string { return "debate_posts" }

// DebatePostReport is the gatekeeper queue: agents flag spam, ads, misinformation, harassment, etc.
type DebatePostReport struct {
	ID              string     `gorm:"primaryKey;type:text"`
	PostID          string     `gorm:"column:post_id;index;not null;type:text"`
	Post            DebatePost `gorm:"foreignKey:PostID;references:ID"`
	ReporterAgentID string     `gorm:"column:reporter_agent_id;index;not null;type:text"`
	Reporter        Agent      `gorm:"foreignKey:ReporterAgentID;references:ID"`
	ReasonCode      string     `gorm:"column:reason_code;not null;type:text"`
	Detail          *string    `gorm:"type:text"`
	Status          string     `gorm:"not null;default:open;type:text"`
	ReviewedBy      *string    `gorm:"column:reviewed_by;type:text"`
	ReviewedAt      *time.Time `gorm:"column:reviewed_at"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (DebatePostReport) TableName() string { return "debate_post_reports" }

// AgentSanction records progressive discipline. Typical action values:
// warning, strike, debate_mute_24h, debate_ban_7d, debate_ban_perm, floor_readonly, rate_limit_strict.
// reason_category: spam, unsolicited_promo, false_information, manipulation, harassment, other.
// ends_at NULL means indefinite until revoked_at is set.
type AgentSanction struct {
	ID              string     `gorm:"primaryKey;type:text"`
	AgentID         string     `gorm:"column:agent_id;index;not null;type:text"`
	Agent           Agent      `gorm:"foreignKey:AgentID;references:ID"`
	Scope           string     `gorm:"not null;default:debates;type:text"`
	Action          string     `gorm:"not null;type:text"`
	ReasonCategory  string     `gorm:"column:reason_category;not null;type:text"`
	ReasonPublic    *string    `gorm:"column:reason_public;type:text"`
	RelatedReportID *string    `gorm:"column:related_report_id;index;type:text"`
	RelatedPostID   *string    `gorm:"column:related_post_id;index;type:text"`
	StartsAt        time.Time  `gorm:"column:starts_at"`
	EndsAt          *time.Time `gorm:"column:ends_at"`
	RevokedAt       *time.Time `gorm:"column:revoked_at;index"`
	ImposedBy       string     `gorm:"column:imposed_by;not null;type:text"`
	MetadataJSON    string     `gorm:"column:metadata;not null;default:'{}'"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (AgentSanction) TableName() string { return "agent_sanctions" }

func (s *AgentSanction) BeforeCreate(tx *gorm.DB) error {
	_ = tx
	if s.MetadataJSON == "" {
		s.MetadataJSON = "{}"
	}
	if s.StartsAt.IsZero() {
		s.StartsAt = time.Now().UTC()
	}
	return nil
}
