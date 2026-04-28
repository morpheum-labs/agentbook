package api

import (
	"time"

	"github.com/morpheumlabs/agentbook/clawgotcha/internal/db"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/events"
	"gorm.io/gorm"
)

// NewSidecarServer builds a Server with DB + event fan-out for non-HTTP callers (e.g. CLI).
func NewSidecarServer(gdb *gorm.DB) *Server {
	return &Server{
		db:  gdb,
		hub: events.NewHub(),
		dispatcher: &events.WebhookDispatcher{
			DB:     gdb,
			Client: events.DefaultHTTPClient(),
		},
	}
}

// UpdateAgentSystemPromptByName updates system_prompt for an agent by name (CLI / compose).
func (s *Server) UpdateAgentSystemPromptByName(agentName, systemPrompt string) error {
	var a db.SwarmAgent
	if err := s.db.Where("name = ?", agentName).First(&a).Error; err != nil {
		return err
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		if err := tx.Model(&a).Updates(map[string]interface{}{
			"system_prompt":    systemPrompt,
			"current_revision": gorm.Expr("current_revision + 1"),
			"last_changed_at":  now,
		}).Error; err != nil {
			return err
		}
		return tx.First(&a, "id = ?", a.ID).Error
	})
	if err != nil {
		return err
	}
	s.emit(events.ChangeEvent{
		EventType:          events.EventAgentUpdated,
		AffectedEntityType: events.EntityAgent,
		AffectedIDs:        []string{a.ID.String()},
		NewRevision:        a.CurrentRevision,
		TS:                 events.NowRFC3339Nano(),
	})
	return nil
}
