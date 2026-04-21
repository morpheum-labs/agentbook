package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
)

func TestFloorCreateTopicProposal_scanner(t *testing.T) {
	s := testServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body := map[string]any{
		"source_kind":       "scanner",
		"selected_event":    "Polymarket X",
		"title":             "Will X resolve?",
		"topic_class":       "Tech",
		"category":          "AI",
		"resolution_rule":   "Official oracle outcome",
		"deadline":          "2026-09-30",
		"source_of_truth":   "contract rules",
		"why_track":         "Liquidity signal",
		"expected_signal":   "Price convergence",
	}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.Post(ts.URL+"/api/v1/floor/topic-proposals", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status %d", res.StatusCode)
	}
	var out map[string]any
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	id, _ := out["id"].(string)
	if id == "" {
		t.Fatalf("missing id: %#v", out)
	}
	if out["status"] != "pending_review" {
		t.Fatalf("status: %#v", out["status"])
	}

	var row dbpkg.FloorTopicProposal
	if err := s.DB.First(&row, "id = ?", id).Error; err != nil {
		t.Fatal(err)
	}
	if row.Title != body["title"] {
		t.Fatalf("title mismatch")
	}
}

func TestFloorCreateTopicProposal_manual(t *testing.T) {
	s := testServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body := map[string]any{
		"source_kind":       "manual",
		"manual_url":        "https://example.com/event/1",
		"title":             "Manual source topic",
		"topic_class":       "Macro",
		"category":          "Rates",
		"resolution_rule":   "Fed statement",
		"deadline":          "2026-12-01",
		"source_of_truth":   "FOMC",
		"why_track":         "Policy path",
		"expected_signal":   "Dot plot shift",
	}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.Post(ts.URL+"/api/v1/floor/topic-proposals", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestFloorCreateTopicProposal_validation(t *testing.T) {
	s := testServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	// scanner without selected_event
	res, err := http.Post(ts.URL+"/api/v1/floor/topic-proposals", "application/json", bytes.NewReader([]byte(`{"source_kind":"scanner","title":"x","category":"c","deadline":"d","resolution_rule":"r","source_of_truth":"s","why_track":"w","expected_signal":"e"}`)))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", res.StatusCode)
	}
}
