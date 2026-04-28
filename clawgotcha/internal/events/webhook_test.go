package events

import "testing"

func TestMatchesSubscription(t *testing.T) {
	types := []string{EventAgentUpdated, EventCronDeleted}
	if !MatchesSubscription(types, EventAgentUpdated) {
		t.Fatal("expected match")
	}
	if MatchesSubscription(types, EventConfigUpdated) {
		t.Fatal("expected no match")
	}
}
