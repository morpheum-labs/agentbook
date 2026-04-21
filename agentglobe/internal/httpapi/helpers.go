package httpapi

import (
	"strings"
	"time"

	"github.com/morpheumlabs/agentbook/agentglobe/internal/db"
)

const onlineWindow = 10 * time.Minute

func agentOnline(a *db.Agent) bool {
	if a == nil || a.LastSeen == nil {
		return false
	}
	return time.Since(*a.LastSeen) < onlineWindow
}

func agentMap(a *db.Agent, includeKey bool) map[string]any {
	display := strings.TrimSpace(a.Name)
	if a.DisplayName != nil && strings.TrimSpace(*a.DisplayName) != "" {
		display = strings.TrimSpace(*a.DisplayName)
	}
	handle := strings.TrimSpace(a.Name)
	if a.FloorHandle != nil && strings.TrimSpace(*a.FloorHandle) != "" {
		handle = strings.TrimSpace(*a.FloorHandle)
	}
	m := map[string]any{
		"id":                 a.ID,
		"name":               a.Name,
		"display_name":       display,
		"handle":             handle,
		"created_at":         a.CreatedAt.UTC().Format(time.RFC3339Nano),
		"registered_at":      a.CreatedAt.UTC().Format(time.RFC3339Nano),
		"platform_verified":  a.PlatformVerified,
	}
	if includeKey {
		m["api_key"] = a.APIKey
	}
	if a.Bio != nil && strings.TrimSpace(*a.Bio) != "" {
		m["bio"] = strings.TrimSpace(*a.Bio)
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
