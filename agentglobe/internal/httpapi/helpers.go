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
	updatedAt := a.UpdatedAt
	if updatedAt.IsZero() && !a.CreatedAt.IsZero() {
		updatedAt = a.CreatedAt
	}
	m := map[string]any{
		"id":                 a.ID,
		"name":               a.Name,
		"display_name":       display,
		"handle":             handle,
		"created_at":         a.CreatedAt.UTC().Format(time.RFC3339Nano),
		"registered_at":      a.CreatedAt.UTC().Format(time.RFC3339Nano),
		"platform_verified":  a.PlatformVerified,
		"updated_at":         updatedAt.UTC().Format(time.RFC3339Nano),
	}
	if includeKey {
		m["api_key"] = a.APIKey
	}
	if a.Bio != nil && strings.TrimSpace(*a.Bio) != "" {
		m["bio"] = strings.TrimSpace(*a.Bio)
	}
	if a.PublicKey != nil && strings.TrimSpace(*a.PublicKey) != "" {
		m["public_key"] = strings.TrimSpace(*a.PublicKey)
	}
	if a.HumanWalletAddress != nil && strings.TrimSpace(*a.HumanWalletAddress) != "" {
		m["human_wallet_address"] = strings.TrimSpace(*a.HumanWalletAddress)
	}
	if a.YoloWalletAddress != nil && strings.TrimSpace(*a.YoloWalletAddress) != "" {
		m["yolo_wallet_address"] = strings.TrimSpace(*a.YoloWalletAddress)
	}
	if a.AvatarURL != nil && strings.TrimSpace(*a.AvatarURL) != "" {
		m["avatar_url"] = strings.TrimSpace(*a.AvatarURL)
	}
	md := a.Metadata()
	if len(md) > 0 {
		m["metadata"] = md
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
