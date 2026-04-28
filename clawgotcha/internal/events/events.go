package events

import (
	"encoding/json"
	"strings"
	"time"
)

// Entity types for change notifications.
const (
	EntityAgent    = "agent"
	EntityCronJob  = "cron_job"
	EntityConfig   = "config"
	EntityInstance = "runtime_instance"
)

// Event type strings (match webhook subscription filters).
const (
	EventAgentUpdated   = "agent.updated"
	EventAgentDeleted   = "agent.deleted"
	EventCronUpdated    = "cron.updated"
	EventCronDeleted    = "cron.deleted"
	EventConfigUpdated  = "config.updated"
	EventInstanceOnline = "runtime.online"
)

// ChangeEvent is broadcast over SSE and webhooks.
type ChangeEvent struct {
	EventType          string   `json:"event_type"`
	AffectedEntityType string   `json:"affected_entity_type"`
	AffectedIDs        []string `json:"affected_ids"`
	NewRevision        int64    `json:"new_revision"`
	TS                 string   `json:"ts"`
}

// NowRFC3339Nano returns UTC timestamp string for events.
func NowRFC3339Nano() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

// MarshalJSONBytes returns compact JSON for signing and SSE.
func MarshalJSONBytes(ev ChangeEvent) ([]byte, error) {
	return json.Marshal(ev)
}

// MatchesSubscription returns true if typ is listed in subscription event_types (exact match).
func MatchesSubscription(eventTypes []string, typ string) bool {
	for _, e := range eventTypes {
		if strings.TrimSpace(e) == typ {
			return true
		}
	}
	return false
}
