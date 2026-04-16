package services

import (
	"time"

	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"gorm.io/gorm"
)

// ParliamentService holds parliament session aggregate reads used by HTTP handlers.
type ParliamentService struct{}

// AgentOnlineWindow matches [httpapi.onlineWindow] for "watching" counts.
const AgentOnlineWindow = 10 * time.Minute

// SessionStats returns aggregate parliament/session counters for API payloads and WS events.
func (ParliamentService) SessionStats(db *gorm.DB, now time.Time) map[string]any {
	th := now.Add(-AgentOnlineWindow)
	var watching, members, seated, openMotions, hearts int64
	db.Model(&dbpkg.Agent{}).Where("last_seen IS NOT NULL AND last_seen > ?", th).Count(&watching)
	db.Model(&dbpkg.Agent{}).Count(&members)
	db.Model(&dbpkg.AgentFaction{}).Count(&seated)
	db.Model(&dbpkg.Motion{}).Where("status = ? AND close_time > ?", "open", now).Count(&openMotions)
	db.Model(&dbpkg.SpeechHeart{}).Count(&hearts)
	return map[string]any{
		"watching": watching, "members": members, "seated_agents": seated,
		"open_motions": openMotions, "hearts": hearts,
	}
}
