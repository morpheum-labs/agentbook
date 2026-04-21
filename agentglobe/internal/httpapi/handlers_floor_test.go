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
		QuestionID:           q.ID,
		DigestDate:           now.Format("2006-01-02"),
		ConsensusLevel:       "consensus",
		Probability:          0.55,
		ProbabilityDelta:     0.01,
		Summary:              "Test digest",
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

	t.Run("topic details alias matches get question", func(t *testing.T) {
		res, err := http.Get(base + "/topics/" + q.ID + "/detail?include=digest")
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
		if m["id"] != q.ID {
			t.Fatalf("id: %v", m["id"])
		}
	})

	t.Run("question digest-history matches digests", func(t *testing.T) {
		for _, path := range []string{
			"/questions/" + q.ID + "/digest-history",
			"/questions/" + q.ID + "/digests",
			"/topics/" + q.ID + "/digest-history",
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

	t.Run("topics page composed payload", func(t *testing.T) {
		res, err := http.Get(base + "/topics")
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("status %d", res.StatusCode)
		}
		var body map[string]any
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		hdr, ok := body["header"].(map[string]any)
		if !ok || hdr["title"] != "Topics" {
			t.Fatalf("header: %#v", body["header"])
		}
		rows, ok := body["browse_rows"].([]any)
		if !ok || len(rows) < 1 {
			t.Fatalf("browse_rows: %#v", body["browse_rows"])
		}
		first, ok := rows[0].(map[string]any)
		if !ok || first["topic_id"] == nil || first["probability_long"] == nil {
			t.Fatalf("first row: %#v", rows[0])
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

func TestFloorIndexPageUsesDBWhenSeeded(t *testing.T) {
	s := testServer(t)
	if err := dbpkg.SeedFloorDemoIndex(s.DB); err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	res, err := http.Get(ts.URL + "/api/v1/floor/index")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	rows, ok := body["rows"].([]any)
	if !ok || len(rows) != 5 {
		t.Fatalf("rows len: %v", body["rows"])
	}
	first, ok := rows[0].(map[string]any)
	if !ok || first["index_id"] != "I.01" {
		t.Fatalf("first row: %#v", rows[0])
	}
	if first["watchlist_locked"] != true {
		t.Fatalf("default tier should lock watchlist, got %#v", first["watchlist_locked"])
	}
	res2, err := http.Get(ts.URL + "/api/v1/floor/index?tier=terminal")
	if err != nil {
		t.Fatal(err)
	}
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("tier status %d", res2.StatusCode)
	}
	var body2 map[string]any
	if err := json.NewDecoder(res2.Body).Decode(&body2); err != nil {
		t.Fatal(err)
	}
	rows2, ok := body2["rows"].([]any)
	if !ok || len(rows2) < 1 {
		t.Fatal(rows2)
	}
	r0, ok := rows2[0].(map[string]any)
	if !ok || r0["watchlist_locked"] != false {
		t.Fatalf("terminal tier should unlock watchlist: %#v", r0["watchlist_locked"])
	}
}

func TestFloorIndexPageJSON(t *testing.T) {
	s := testServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	res, err := http.Get(ts.URL + "/api/v1/floor/index")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	rows, ok := body["rows"].([]any)
	if !ok || len(rows) < 1 {
		t.Fatalf("rows: %v", body["rows"])
	}
	first, ok := rows[0].(map[string]any)
	if !ok || first["index_id"] == nil {
		t.Fatalf("first row: %#v", rows[0])
	}
	if body["selected_index"] == nil {
		t.Fatal("expected selected_index")
	}
	if body["summary_chips"] == nil {
		t.Fatal("expected summary_chips")
	}
}

func TestFloorTopicsPageUsesDBWhenDemoSeeded(t *testing.T) {
	s := testServer(t)
	if err := dbpkg.SeedFloorDemoTopics(s.DB); err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	res, err := http.Get(ts.URL + "/api/v1/floor/topics")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	rows, ok := body["browse_rows"].([]any)
	if !ok || len(rows) != 4 {
		t.Fatalf("browse_rows len: %v", body["browse_rows"])
	}
	first, ok := rows[0].(map[string]any)
	if !ok || first["topic_id"] == nil || first["open_topic_details_url"] == nil {
		t.Fatalf("first browse row: %#v", rows[0])
	}
	if body["selected_topic"] == nil {
		t.Fatal("expected selected_topic")
	}
}

func TestFloorDiscoverPageEmpty(t *testing.T) {
	s := testServer(t)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	res, err := http.Get(ts.URL + "/api/v1/floor/discover")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	ranked, _ := body["ranked"].([]any)
	if ranked == nil {
		t.Fatalf("ranked: %#v", body["ranked"])
	}
	if len(ranked) != 0 {
		t.Fatalf("want empty ranked, got %d", len(ranked))
	}
}

func TestFloorDiscoverPageWithDemoSeed(t *testing.T) {
	s := testServer(t)
	if err := dbpkg.SeedFloorDemoTopics(s.DB); err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	res, err := http.Get(ts.URL + "/api/v1/floor/discover")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	ranked, ok := body["ranked"].([]any)
	if !ok || len(ranked) < 1 {
		t.Fatalf("ranked: %#v", body["ranked"])
	}
	first, ok := ranked[0].(map[string]any)
	if !ok {
		t.Fatalf("first ranked row type %T", ranked[0])
	}
	if first["id"] != "floor-demo-agent-omega" {
		t.Fatalf("expected omega first ranked, got id=%v", first["id"])
	}
	emerging, ok := body["emerging"].([]any)
	if !ok || len(emerging) < 1 {
		t.Fatalf("emerging: %#v", body["emerging"])
	}
	unq, ok := body["unqualified"].([]any)
	if !ok || len(unq) < 1 {
		t.Fatalf("unqualified: %#v", body["unqualified"])
	}
}
