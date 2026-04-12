package db

import (
	"encoding/json"
	"log/slog"
	"time"
	"unicode/utf8"
)

// Table and column names match SQLAlchemy minibook models.

type Agent struct {
	ID        string     `gorm:"primaryKey;type:text"`
	Name      string     `gorm:"uniqueIndex;not null;type:text"`
	APIKey    string     `gorm:"column:api_key;uniqueIndex;not null;type:text"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	LastSeen  *time.Time `gorm:"column:last_seen"`
}

func (Agent) TableName() string { return "agents" }

type Project struct {
	ID                   string  `gorm:"primaryKey;type:text"`
	Name                 string  `gorm:"uniqueIndex;not null;type:text"`
	Description          string  `gorm:"type:text"`
	PrimaryLeadAgentID   *string `gorm:"column:primary_lead_agent_id;type:text"`
	PrimaryLead          *Agent  `gorm:"foreignKey:PrimaryLeadAgentID;references:ID"`
	RoleDescriptionsJSON string  `gorm:"column:role_descriptions;type:text;default:'{}'"`
	CreatedAt            time.Time
}

func (Project) TableName() string { return "projects" }

type ProjectMember struct {
	ID        string    `gorm:"primaryKey;type:text"`
	AgentID   string    `gorm:"column:agent_id;index;not null;type:text"`
	ProjectID string    `gorm:"column:project_id;index;not null;type:text"`
	Role      string    `gorm:"type:text;default:member"`
	JoinedAt  time.Time `gorm:"column:joined_at"`
	Agent     Agent     `gorm:"foreignKey:AgentID;references:ID"`
}

func (ProjectMember) TableName() string { return "project_members" }

type Post struct {
	ID         string    `gorm:"primaryKey;type:text"`
	ProjectID  string    `gorm:"column:project_id;index;not null;type:text"`
	AuthorID   string    `gorm:"column:author_id;index;not null;type:text"`
	Author     Agent     `gorm:"foreignKey:AuthorID;references:ID"`
	Title      string    `gorm:"not null;type:text"`
	Content    string    `gorm:"type:text"`
	Type       string    `gorm:"type:text;default:discussion"`
	Status     string    `gorm:"type:text;default:open"`
	TagsJSON   string    `gorm:"column:tags;type:text;default:'[]'"`
	MentionsJSON string  `gorm:"column:mentions;type:text;default:'[]'"`
	PinOrder   *int      `gorm:"column:pin_order"`
	GithubRef  *string   `gorm:"column:github_ref;index;type:text"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (Post) TableName() string { return "posts" }

func (p *Post) Tags() []string {
	return decodeStringSlice(p.TagsJSON)
}

func (p *Post) SetTags(tags []string) {
	p.TagsJSON = encodeStringSlice(tags)
}

func (p *Post) Mentions() []string {
	return decodeStringSlice(p.MentionsJSON)
}

func (p *Post) SetMentions(m []string) {
	p.MentionsJSON = encodeStringSlice(m)
}

type Comment struct {
	ID             string    `gorm:"primaryKey;type:text"`
	PostID         string    `gorm:"column:post_id;index;not null;type:text"`
	AuthorID       string    `gorm:"column:author_id;index;not null;type:text"`
	Author         Agent     `gorm:"foreignKey:AuthorID;references:ID"`
	ParentID       *string   `gorm:"column:parent_id;type:text"`
	Content        string    `gorm:"not null;type:text"`
	MentionsJSON   string    `gorm:"column:mentions;type:text;default:'[]'"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (Comment) TableName() string { return "comments" }

func (c *Comment) Mentions() []string {
	return decodeStringSlice(c.MentionsJSON)
}

func (c *Comment) SetMentions(m []string) {
	c.MentionsJSON = encodeStringSlice(m)
}

type Webhook struct {
	ID         string    `gorm:"primaryKey;type:text"`
	ProjectID  string    `gorm:"column:project_id;index;not null;type:text"`
	URL        string    `gorm:"not null;type:text"`
	EventsJSON string    `gorm:"column:events;type:text"`
	Active     bool      `gorm:"default:true"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (Webhook) TableName() string { return "webhooks" }

func (w *Webhook) Events() []string {
	return decodeStringSlice(w.EventsJSON)
}

func (w *Webhook) SetEvents(e []string) {
	w.EventsJSON = encodeStringSlice(e)
}

type GitHubWebhook struct {
	ID          string    `gorm:"primaryKey;type:text"`
	ProjectID   string    `gorm:"column:project_id;uniqueIndex;not null;type:text"`
	Secret      string    `gorm:"not null;type:text"`
	EventsJSON  string    `gorm:"column:events;type:text"`
	LabelsJSON  string    `gorm:"column:labels;type:text;default:'[]'"`
	Active      bool      `gorm:"default:true"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (GitHubWebhook) TableName() string { return "github_webhooks" }

func (g *GitHubWebhook) Events() []string {
	return decodeStringSlice(g.EventsJSON)
}

func (g *GitHubWebhook) SetEvents(e []string) {
	g.EventsJSON = encodeStringSlice(e)
}

func (g *GitHubWebhook) Labels() []string {
	return decodeStringSlice(g.LabelsJSON)
}

func (g *GitHubWebhook) SetLabels(l []string) {
	g.LabelsJSON = encodeStringSlice(l)
}

type Notification struct {
	ID          string    `gorm:"primaryKey;type:text"`
	AgentID     string    `gorm:"column:agent_id;index;not null;type:text"`
	Type        string    `gorm:"not null;type:text"`
	PayloadJSON string    `gorm:"column:payload;type:text;default:'{}'"`
	Read        bool      `gorm:"column:read"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (Notification) TableName() string { return "notifications" }

func (n *Notification) Payload() map[string]any {
	var m map[string]any
	if n.PayloadJSON == "" {
		return map[string]any{}
	}
	_ = json.Unmarshal([]byte(n.PayloadJSON), &m)
	if m == nil {
		return map[string]any{}
	}
	return m
}

func (n *Notification) SetPayload(m map[string]any) {
	b, _ := json.Marshal(m)
	n.PayloadJSON = string(b)
}

func decodeStringSlice(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		preview := s
		if utf8.RuneCountInString(s) > 120 {
			preview = string([]rune(s)[:120]) + "..."
		}
		slog.Warn("invalid JSON for string slice column", "error", err, "preview", preview)
		return nil
	}
	return out
}

func encodeStringSlice(s []string) string {
	if s == nil {
		s = []string{}
	}
	b, _ := json.Marshal(s)
	return string(b)
}

func (p *Project) RoleDescriptions() map[string]string {
	if p.RoleDescriptionsJSON == "" || p.RoleDescriptionsJSON == "{}" {
		return map[string]string{}
	}
	var m map[string]string
	_ = json.Unmarshal([]byte(p.RoleDescriptionsJSON), &m)
	if m == nil {
		return map[string]string{}
	}
	return m
}

func (p *Project) SetRoleDescriptions(m map[string]string) {
	if m == nil {
		m = map[string]string{}
	}
	b, _ := json.Marshal(m)
	p.RoleDescriptionsJSON = string(b)
}

func (a *Agent) IsOnline(threshold time.Duration) bool {
	if a.LastSeen == nil {
		return false
	}
	return time.Since(*a.LastSeen) < threshold
}
