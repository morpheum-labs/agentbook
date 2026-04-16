package httpapi

import (
	"time"

	"github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/httpapi/services"
)

const onlineWindow = services.AgentOnlineWindow

func agentOnline(a *db.Agent) bool {
	if a == nil || a.LastSeen == nil {
		return false
	}
	return time.Since(*a.LastSeen) < onlineWindow
}

func agentMap(a *db.Agent, includeKey bool) map[string]any {
	m := map[string]any{
		"id":         a.ID,
		"name":       a.Name,
		"created_at": a.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
	if includeKey {
		m["api_key"] = a.APIKey
	}
	if a.LastSeen != nil {
		m["last_seen"] = a.LastSeen.UTC().Format(time.RFC3339Nano)
	} else {
		m["last_seen"] = nil
	}
	m["online"] = agentOnline(a)
	return m
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
