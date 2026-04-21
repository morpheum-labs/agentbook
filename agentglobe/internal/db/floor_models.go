package db

import "time"

// Floor models map to floor_* tables (AgentFloor). GORM AutoMigrate is the source of truth for column types.

type FloorQuestion struct {
	ID                   string    `gorm:"primaryKey;type:text"`
	Title                string    `gorm:"not null;type:text"`
	Category             string    `gorm:"not null;type:text"`
	ResolutionCondition  string    `gorm:"column:resolution_condition;not null;type:text"`
	Deadline             string    `gorm:"not null;type:text"`
	Probability          float64   `gorm:"not null"`
	ProbabilityDelta     float64   `gorm:"column:probability_delta;not null"`
	AgentCount           int       `gorm:"column:agent_count;not null"`
	StakedCount          int       `gorm:"column:staked_count;not null"`
	Status               string    `gorm:"not null;type:text;default:open"`
	ClusterBreakdownJSON string    `gorm:"column:cluster_breakdown_json;not null;type:text;default:'{}'"`
	ZkVerifiedPct        *float64  `gorm:"column:zk_verified_pct"`
	WmContextID          *string   `gorm:"column:wm_context_id;type:text"`
	CreatedAt            time.Time `gorm:"column:created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at"`
}

func (FloorQuestion) TableName() string { return "floor_questions" }

type FloorPosition struct {
	ID                    string        `gorm:"primaryKey;type:text"`
	QuestionID            string        `gorm:"column:question_id;index;not null;type:text"`
	Question              FloorQuestion `gorm:"foreignKey:QuestionID;references:ID"`
	AgentID               string        `gorm:"column:agent_id;index;not null;type:text"`
	Agent                 Agent         `gorm:"foreignKey:AgentID;references:ID"`
	Direction             string        `gorm:"not null;type:text"`
	StakedAt              time.Time     `gorm:"column:staked_at"`
	Body                  string        `gorm:"not null;type:text;default:''"`
	Language              string        `gorm:"not null;type:text;default:'EN'"`
	AccuracyScoreAtStake  *float64      `gorm:"column:accuracy_score_at_stake"`
	InferenceProof        *string       `gorm:"column:inference_proof;type:text"`
	ProofType             *string       `gorm:"column:proof_type;type:text"`
	RegionalCluster       *string       `gorm:"column:regional_cluster;type:text"`
	Resolved              bool          `gorm:"not null"`
	Outcome               string        `gorm:"not null;type:text;default:pending"`
	ChallengeOpen         bool          `gorm:"column:challenge_open;not null"`
	SourcePostID          *string       `gorm:"column:source_post_id;type:text"`
	SourceCommentID       *string       `gorm:"column:source_comment_id;type:text"`
	ExternalSignalIDsJSON string        `gorm:"column:external_signal_ids_json;not null;type:text;default:'[]'"`
	CreatedAt             time.Time     `gorm:"column:created_at"`
}

func (FloorPosition) TableName() string { return "floor_positions" }

// FloorExternalSignal caches World Monitor (or other) OSINT payloads tied to a floor question (F7 audit trail).
type FloorExternalSignal struct {
	ID                   string    `gorm:"primaryKey;type:text"`
	QuestionID           *string   `gorm:"column:question_id;index;type:text"`
	TopicClass           *string   `gorm:"column:topic_class;type:text"`
	FetchedAt            time.Time `gorm:"column:fetched_at;index"`
	Source               string    `gorm:"not null;type:text;default:worldmonitor"`
	RawDataJSON          string    `gorm:"column:raw_data_json;not null;type:text;default:'{}'"`
	InstabilityIndexJSON string    `gorm:"column:instability_index_json;not null;type:text;default:'{}'"`
	GeoConvergenceJSON   string    `gorm:"column:geo_convergence_json;not null;type:text;default:'{}'"`
	ForecastSummaryJSON  string    `gorm:"column:forecast_summary_json;not null;type:text;default:'{}'"`
	UpstreamSignatureMs  *int64    `gorm:"column:upstream_signature_ms"`
	FetchError           *string   `gorm:"column:fetch_error;type:text"`
}

func (FloorExternalSignal) TableName() string { return "floor_external_signals" }

type FloorAgentTopicStat struct {
	AgentID    string    `gorm:"primaryKey;column:agent_id;type:text"`
	TopicClass string    `gorm:"primaryKey;column:topic_class;type:text"`
	Calls      int       `gorm:"not null"`
	Correct    int       `gorm:"not null"`
	Score      float64   `gorm:"not null"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (FloorAgentTopicStat) TableName() string { return "floor_agent_topic_stats" }

type FloorAgentInferenceProfile struct {
	AgentID           string    `gorm:"primaryKey;column:agent_id;type:text"`
	InferenceVerified bool      `gorm:"column:inference_verified;not null"`
	ProofType         *string   `gorm:"column:proof_type;type:text"`
	CredentialPath    *string   `gorm:"column:credential_path;type:text"`
	UpdatedAt         time.Time `gorm:"column:updated_at"`
}

func (FloorAgentInferenceProfile) TableName() string { return "floor_agent_inference_profile" }

type FloorDigestEntry struct {
	ID                   string    `gorm:"primaryKey;type:text"`
	QuestionID           string    `gorm:"column:question_id;not null;type:text;uniqueIndex:floor_digest_question_date"`
	DigestDate           string    `gorm:"column:digest_date;not null;type:text;uniqueIndex:floor_digest_question_date"`
	ConsensusLevel       string    `gorm:"column:consensus_level;not null;type:text"`
	Probability          float64   `gorm:"not null"`
	ProbabilityDelta     float64   `gorm:"column:probability_delta;not null"`
	Summary              string    `gorm:"not null;type:text"`
	TopLongAgentID       *string   `gorm:"column:top_long_agent_id;type:text"`
	TopShortAgentID      *string   `gorm:"column:top_short_agent_id;type:text"`
	ClusterBreakdownJSON string    `gorm:"column:cluster_breakdown_json;not null;type:text;default:'{}'"`
	LlmIndexHits         *int      `gorm:"column:llm_index_hits"`
	CreatedAt            time.Time `gorm:"column:created_at"`
}

func (FloorDigestEntry) TableName() string { return "floor_digest_entries" }

type FloorQuestionProbabilityPoint struct {
	ID          string    `gorm:"primaryKey;type:text"`
	QuestionID  string    `gorm:"column:question_id;index;not null;type:text"`
	CapturedAt  time.Time `gorm:"column:captured_at"`
	Probability float64   `gorm:"not null"`
	Source      string    `gorm:"not null;type:text;default:aggregate"`
}

func (FloorQuestionProbabilityPoint) TableName() string { return "floor_question_probability_points" }

type FloorShieldClaim struct {
	ID                    string                 `gorm:"primaryKey;type:text"`
	Keyword               string                 `gorm:"not null;type:text"`
	AgentID               string                 `gorm:"column:agent_id;index;not null;type:text"`
	Agent                 Agent                  `gorm:"foreignKey:AgentID;references:ID"`
	Category              *string                `gorm:"type:text"`
	Rationale             string                 `gorm:"not null;type:text;default:''"`
	StakedAt              time.Time              `gorm:"column:staked_at"`
	ChallengePeriodEndsAt *time.Time             `gorm:"column:challenge_period_ends_at"`
	AccuracyThresholdMet  bool                   `gorm:"column:accuracy_threshold_met;not null"`
	ChallengeCount        int                    `gorm:"column:challenge_count;not null"`
	ChallengePeriodOpen   bool                   `gorm:"column:challenge_period_open;not null"`
	Sustained             bool                   `gorm:"not null"`
	DigestPublished       bool                   `gorm:"column:digest_published;not null"`
	InferenceProof        *string                `gorm:"column:inference_proof;type:text"`
	StrengthScore         *float64               `gorm:"column:strength_score"`
	Status                string                 `gorm:"not null;type:text;default:active"`
	LinkedQuestionID      *string                `gorm:"column:linked_question_id;type:text"`
	CreatedAt             time.Time              `gorm:"column:created_at"`
	UpdatedAt             time.Time              `gorm:"column:updated_at"`
	Challenges            []FloorShieldChallenge `gorm:"foreignKey:ClaimID;references:ID"`
}

func (FloorShieldClaim) TableName() string { return "floor_shield_claims" }

type FloorShieldChallenge struct {
	ID                string                     `gorm:"primaryKey;type:text"`
	ClaimID           string                     `gorm:"column:claim_id;index;not null;type:text"`
	ChallengerAgentID string                     `gorm:"column:challenger_agent_id;index;not null;type:text"`
	Challenger        Agent                      `gorm:"foreignKey:ChallengerAgentID;references:ID"`
	OpenedAt          time.Time                  `gorm:"column:opened_at"`
	ClosesAt          time.Time                  `gorm:"column:closes_at"`
	Resolution        *string                    `gorm:"type:text"`
	ResolvedAt        *time.Time                 `gorm:"column:resolved_at"`
	TallyJSON         string                     `gorm:"column:tally_json;not null;type:text;default:'{}'"`
	Votes             []FloorShieldChallengeVote `gorm:"foreignKey:ChallengeID;references:ID"`
}

func (FloorShieldChallenge) TableName() string { return "floor_shield_challenges" }

type FloorShieldChallengeVote struct {
	ID           string    `gorm:"primaryKey;type:text"`
	ChallengeID  string    `gorm:"column:challenge_id;not null;type:text;uniqueIndex:floor_shield_vote_chal_voter"`
	VoterAgentID string    `gorm:"column:voter_agent_id;not null;type:text;uniqueIndex:floor_shield_vote_chal_voter"`
	Voter        Agent     `gorm:"foreignKey:VoterAgentID;references:ID"`
	Vote         string    `gorm:"not null;type:text"`
	Weight       float64   `gorm:"not null;default:1"`
	CastAt       time.Time `gorm:"column:cast_at"`
}

func (FloorShieldChallengeVote) TableName() string { return "floor_shield_challenge_votes" }

type FloorPositionChallenge struct {
	ID                string     `gorm:"primaryKey;type:text"`
	PositionID        string     `gorm:"column:position_id;index;not null;type:text"`
	ChallengerAgentID string     `gorm:"column:challenger_agent_id;index;not null;type:text"`
	Challenger        Agent      `gorm:"foreignKey:ChallengerAgentID;references:ID"`
	Status            string     `gorm:"not null;type:text;default:open"`
	OpenedAt          time.Time  `gorm:"column:opened_at"`
	ResolvedAt        *time.Time `gorm:"column:resolved_at"`
	ResolutionNotes   *string    `gorm:"column:resolution_notes;type:text"`
}

func (FloorPositionChallenge) TableName() string { return "floor_position_challenges" }

type FloorResearchArticle struct {
	ID              string    `gorm:"primaryKey;type:text"`
	Title           string    `gorm:"not null;type:text"`
	Summary         string    `gorm:"not null;type:text"`
	Body            *string   `gorm:"type:text"`
	ClusterTagsJSON string    `gorm:"column:cluster_tags_json;not null;type:text;default:'[]'"`
	PublishedAt     *string   `gorm:"column:published_at;type:text"`
	DigestDate      *string   `gorm:"column:digest_date;type:text"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
}

func (FloorResearchArticle) TableName() string { return "floor_research_articles" }

type FloorBroadcast struct {
	ID              string     `gorm:"primaryKey;type:text"`
	Title           string     `gorm:"not null;type:text"`
	Status          string     `gorm:"not null;type:text;default:scheduled"`
	StartsAt        time.Time  `gorm:"column:starts_at"`
	EndsAt          *time.Time `gorm:"column:ends_at"`
	QuestionIDsJSON string     `gorm:"column:question_ids_json;not null;type:text;default:'[]'"`
	ArchiveURL      *string    `gorm:"column:archive_url;type:text"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
}

func (FloorBroadcast) TableName() string { return "floor_broadcasts" }

// FloorIndexPageMeta is a singleton row (id FloorIndexPageMetaDefaultID) for AgentFloor GET /floor/index header, chips, filters, and lower strip.
type FloorIndexPageMeta struct {
	ID                      string `gorm:"primaryKey;type:text"`
	HeaderTitle             string `gorm:"column:header_title;not null;type:text"`
	HeaderSubtitle          string `gorm:"column:header_subtitle;not null;type:text"`
	HeaderWatchlistTierHint string `gorm:"column:header_watchlist_tier_hint;type:text"`
	SummaryChipsJSON        string `gorm:"column:summary_chips_json;not null;type:text;default:'[]'"`
	FiltersJSON             string `gorm:"column:filters_json;not null;type:text;default:'[]'"`
	LowerStripJSON          string `gorm:"column:lower_strip_json;not null;type:text;default:'{}'"`
	SelectedIndexID         string `gorm:"column:selected_index_id;not null;type:text"`
}

func (FloorIndexPageMeta) TableName() string { return "floor_index_page_meta" }

// FloorIndexPageMetaDefaultID is the primary key for the composed index page configuration row.
const FloorIndexPageMetaDefaultID = "default"

// FloorIndexEntry is one directory row plus its detail panel payload for GET /floor/index.
type FloorIndexEntry struct {
	IndexID              string `gorm:"column:index_id;primaryKey;type:text"`
	SortOrder            int    `gorm:"column:sort_order;not null"`
	Title                string `gorm:"not null;type:text"`
	Type                 string `gorm:"not null;type:text"`
	SignalLabel          string `gorm:"column:signal_label;not null;type:text"`
	ConfidenceLabel      string `gorm:"column:confidence_label;type:text"`
	AccessTier           string `gorm:"column:access_tier;not null;type:text"`
	OpenDetailURL        string `gorm:"column:open_detail_url;not null;type:text"`
	CanWatchlist         bool   `gorm:"column:can_watchlist;not null"`
	Watchlisted          bool   `gorm:"column:watchlisted;not null"`
	Subtitle             string `gorm:"type:text"`
	WhyItMatters         string `gorm:"column:why_it_matters;type:text"`
	CurrentReading       string `gorm:"column:current_reading;type:text"`
	TrustSnapshotJSON    string `gorm:"column:trust_snapshot_json;not null;type:text;default:'{}'"`
	SourceProvenanceJSON string `gorm:"column:source_provenance_json;not null;type:text;default:'{}'"`
	UpdateLogJSON        string `gorm:"column:update_log_json;not null;type:text;default:'[]'"`
}

func (FloorIndexEntry) TableName() string { return "floor_index_entries" }
