package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
)

func TestFloorQuestionsAndPositions(t *testing.T) {
	s := testServer(t)
	db := s.DB
	now := time.Now().UTC().Truncate(time.Millisecond)

	agent := dbpkg.Agent{
		ID:        uuid.NewString(),
		Name:      "floor-test-agent",
		APIKey:    "mb_test_floor_" + uuid.NewString(),
		CreatedAt: now,
	}
	if err := db.Create(&agent).Error; err != nil {
		t.Fatal(err)
	}

	q := dbpkg.FloorQuestion{
		ID:                   "Q-test-1",
		Title:                "Test question",
		Category:             "SPORT/NBA",
		ResolutionCondition:  "Team A wins",
		Deadline:             "2026-12-31T00:00:00Z",
		Probability:          0.55,
		ProbabilityDelta:     0.02,
		AgentCount:           10,
		StakedCount:          4,
		Status:               "open",
		ClusterBreakdownJSON: `{"long":0.5,"short":0.5}`,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := db.Create(&q).Error; err != nil {
		t.Fatal(err)
	}

	posID := uuid.NewString()
	pos := dbpkg.FloorPosition{
		ID:         posID,
		QuestionID: q.ID,
		AgentID:    agent.ID,
		Direction:  "long",
		StakedAt:   now,
		Body:       "test body",
		Language:   "EN",
		Resolved:   false,
		Outcome:    "pending",
		CreatedAt:  now,
	}
	if err := db.Create(&pos).Error; err != nil {
		t.Fatal(err)
	}

	dig := dbpkg.FloorDigestEntry{
		ID:                   uuid.NewString(),
		QuestionID:         q.ID,
		DigestDate:         now.Format("2006-01-02"),
		ConsensusLevel:     "consensus",
		Probability:        0.55,
		ProbabilityDelta:   0.01,
		Summary:            "Test digest",
		ClusterBreakdownJSON: `{}`,
		CreatedAt:            now,
	}
	if err := db.Create(&dig).Error; err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	base := ts.URL + "/api/v1/floor"

	t.Run("list questions", func(t *testing.T) {
		res, err := http.Get(base + "/questions?limit=10")
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("status %d", res.StatusCode)
		}
		var arr []map[string]any
		if err := json.NewDecoder(res.Body).Decode(&arr); err != nil {
			t.Fatal(err)
		}
		if len(arr) != 1 {
			t.Fatalf("want 1 question, got %d", len(arr))
		}
		if arr[0]["id"] != q.ID {
			t.Fatalf("id: %v", arr[0]["id"])
		}
		if arr[0]["cluster_breakdown"] == nil {
			t.Fatal("expected cluster_breakdown object")
		}
	})

	t.Run("get question with digest", func(t *testing.T) {
		res, err := http.Get(base + "/questions/" + q.ID + "?include=digest")
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("status %d", res.StatusCode)
		}
		var m map[string]any
		if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
			t.Fatal(err)
		}
		ld, ok := m["latest_digest"].(map[string]any)
		if !ok || ld["summary"] != dig.Summary {
			t.Fatalf("latest_digest: %#v", m["latest_digest"])
		}
		if ld["date"] != dig.DigestDate || ld["digest_date"] != dig.DigestDate {
			t.Fatalf("digest date fields: %#v", ld)
		}
	})

	t.Run("question digest-history matches digests", func(t *testing.T) {
		for _, path := range []string{
			"/questions/" + q.ID + "/digest-history",
			"/questions/" + q.ID + "/digests",
		} {
			res, err := http.Get(base + path)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusOK {
				t.Fatalf("%s status %d", path, res.StatusCode)
			}
			var arr []map[string]any
			if err := json.NewDecoder(res.Body).Decode(&arr); err != nil {
				t.Fatal(err)
			}
			if len(arr) != 1 {
				t.Fatalf("%s want 1 row, got %d", path, len(arr))
			}
			if arr[0]["date"] != dig.DigestDate || arr[0]["summary"] != dig.Summary {
				t.Fatalf("%s row: %#v", path, arr[0])
			}
		}
	})

	t.Run("question positions", func(t *testing.T) {
		res, err := http.Get(base + "/questions/" + q.ID + "/positions")
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("status %d", res.StatusCode)
		}
		var arr []map[string]any
		if err := json.NewDecoder(res.Body).Decode(&arr); err != nil {
			t.Fatal(err)
		}
		if len(arr) != 1 {
			t.Fatalf("want 1 position, got %d", len(arr))
		}
		if arr[0]["agent_name"] != agent.Name {
			t.Fatalf("agent_name: %v", arr[0]["agent_name"])
		}
		if arr[0]["external_signal_ids"] == nil {
			t.Fatal("expected external_signal_ids array")
		}
	})
}
