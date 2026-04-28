package api

import (
	"fmt"
	"time"

	"github.com/morpheumlabs/agentbook/clawgotcha/internal/db"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/prompt"
)

// swarmAgentResponse is the public JSON for agents: mixed PascalCase (GORM/legacy) plus
// snake_case modular prompt fields (IDENTITY, SOUL, USER.md → identity, soul, user_context).
type swarmAgentResponse struct {
	ID              string     `json:"ID"`
	Name            string     `json:"Name"`
	SystemPrompt    string     `json:"SystemPrompt"`
	Identity        string     `json:"identity"`
	Soul            string     `json:"soul"`
	UserContext     string     `json:"user_context"`
	ModularPrompt   bool       `json:"modular_prompt"`
	Tools           []string   `json:"Tools"`
	Provider        string     `json:"Provider"`
	Model           string     `json:"Model"`
	TimeoutSeconds  int        `json:"TimeoutSeconds"`
	AutonomyLevel   string     `json:"AutonomyLevel"`
	CurrentRevision int64      `json:"current_revision"`
	LastChangedAt   *time.Time `json:"last_changed_at,omitempty"`
	Deleted         bool       `json:"deleted,omitempty"`
	CreatedAt       time.Time  `json:"CreatedAt"`
	UpdatedAt       time.Time  `json:"UpdatedAt"`
}

func toSwarmAgentResponse(a db.SwarmAgent) swarmAgentResponse {
	base := swarmAgentResponse{
		ID:              a.ID.String(),
		Name:            a.Name,
		SystemPrompt:    a.SystemPrompt,
		CurrentRevision: a.CurrentRevision,
		LastChangedAt:   a.LastChangedAt,
		Deleted:         a.DeletedAt.Valid,
		Tools:           a.Tools,
		Provider:        a.Provider,
		Model:           a.Model,
		TimeoutSeconds:  a.TimeoutSeconds,
		AutonomyLevel:   a.AutonomyLevel,
		CreatedAt:       a.CreatedAt,
		UpdatedAt:       a.UpdatedAt,
	}
	p, err := prompt.ParseSections(a.SystemPrompt)
	if err == nil {
		base.Identity = p.Identity
		base.Soul = p.Soul
		base.UserContext = p.User
		base.ModularPrompt = true
		return base
	}
	base.Identity = ""
	base.Soul = ""
	base.UserContext = a.SystemPrompt
	base.ModularPrompt = false
	return base
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// modularKeysPresent is true if the client sent at least one of identity / soul / user_context in the JSON.
func (b *createAgentBody) modularKeysPresent() bool {
	return b.Identity != nil || b.Soul != nil || b.UserContext != nil
}

// resolvedSystemPrompt assembles a stored system_prompt. If any modular field pointer is
// set, the body is in modular mode and the three (nil → "") are combined with MiroClaw
// markers. Otherwise the flat system_prompt string is used.
func (b *createAgentBody) resolvedSystemPrompt() string {
	if b.modularKeysPresent() {
		p := prompt.Pack{
			Identity: derefString(b.Identity),
			Soul:     derefString(b.Soul),
			User:     derefString(b.UserContext),
		}
		return p.CombinedString()
	}
	return b.SystemPrompt
}

// applyPatchSystemPrompt returns (newValue, true, nil) if system_prompt should be updated, or ("", false, nil) if not.
// Modular keys take precedence over system_prompt when any modular pointer is set.
func applyPatchSystemPrompt(current string, b *patchAgentBody) (string, bool, error) {
	modIn := b.Identity != nil || b.Soul != nil || b.UserContext != nil
	if !modIn {
		if b.SystemPrompt != nil {
			return *b.SystemPrompt, true, nil
		}
		return "", false, nil
	}
	allThree := b.Identity != nil && b.Soul != nil && b.UserContext != nil
	p, err := prompt.ParseSections(current)
	if err == nil {
		if b.Identity != nil {
			p.Identity = *b.Identity
		}
		if b.Soul != nil {
			p.Soul = *b.Soul
		}
		if b.UserContext != nil {
			p.User = *b.UserContext
		}
		return p.CombinedString(), true, nil
	}
	if allThree {
		np := prompt.Pack{Identity: derefString(b.Identity), Soul: derefString(b.Soul), User: derefString(b.UserContext)}
		return np.CombinedString(), true, nil
	}
	return "", false, fmt.Errorf("cannot update individual identity/soul/user_context when system_prompt is not in modular form; send all three parts together, or set system_prompt to replace the full prompt")
}
