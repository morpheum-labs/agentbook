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

func TestFloorWorldMonitorContextMockUpstream(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/intelligence/v1/get-risk-scores", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-WorldMonitor-Key") != "wm-test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ciiScores":[{"region":"MENA","combinedScore":88.2,"components":{"geoConvergence":91.4},"computedAt":1710000000000}],"strategicRisks":[]}`))
	})
	mux.HandleFunc("/api/forecast/v1/get-forecasts", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-WorldMonitor-Key") != "wm-test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"forecasts":[{"id":"f1","region":"MENA","title":"Red Sea disruption","probability":0.8,"timeHorizon":"7d","scenario":"base"}],"generatedAt":1710000000001}`))
	})
	wmSrv := httptest.NewServer(mux)
	defer wmSrv.Close()

	t.Setenv("WORLDMONITOR_API_BASE", wmSrv.URL)
	t.Setenv("WORLDMONITOR_API_KEY", "wm-test-key")

	s := testServer(t)
	db := s.DB
	now := time.Now().UTC().Truncate(time.Millisecond)

	agent := dbpkg.Agent{
		ID:        uuid.NewString(),
		Name:      "wm-test-agent",
		APIKey:    "mb_wm_test_" + uuid.NewString(),
		CreatedAt: now,
	}
	if err := db.Create(&agent).Error; err != nil {
		t.Fatal(err)
	}

	q := dbpkg.FloorQuestion{
		ID:                   "Q-wm-1",
		Title:                "MENA stress test",
		Category:             "MACRO/MENA",
		ResolutionCondition:  "Resolved by committee",
		Deadline:             "2026-12-31T00:00:00Z",
		Probability:          0.5,
		ProbabilityDelta:     0,
		AgentCount:           1,
		StakedCount:          0,
		Status:               "open",
		ClusterBreakdownJSON: `{"long":0.5,"neutral":0.1,"short":0.4}`,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := db.Create(&q).Error; err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	base := ts.URL + "/api/v1/floor/questions/" + q.ID + "/context/worldmonitor"

	req, err := http.NewRequest(http.MethodGet, base, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+agent.APIKey)
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("first fetch status %d", res.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	ctx, ok := body["context"].(map[string]any)
	if !ok {
		t.Fatalf("context: %#v", body["context"])
	}
	wm, ok := ctx["worldmonitor"].(map[string]any)
	if !ok {
		t.Fatalf("worldmonitor: %#v", ctx["worldmonitor"])
	}
	inst, ok := wm["instability"].(map[string]any)
	if !ok || inst["MENA"] != float64(88) {
		t.Fatalf("instability MENA: %#v", inst)
	}

	var n int64
	if err := db.Model(&dbpkg.FloorExternalSignal{}).Where("question_id = ?", q.ID).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected 1 cached signal, got %d", n)
	}

	req2, _ := http.NewRequest(http.MethodGet, base, nil)
	req2.Header.Set("Authorization", "Bearer "+agent.APIKey)
	res2, err := ts.Client().Do(req2)
	if err != nil {
		t.Fatal(err)
	}
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("second fetch status %d", res2.StatusCode)
	}
	var body2 map[string]any
	if err := json.NewDecoder(res2.Body).Decode(&body2); err != nil {
		t.Fatal(err)
	}
	if body2["cache_hit"] != true {
		t.Fatalf("expected cache_hit true, got %#v", body2["cache_hit"])
	}
}

func TestFloorWorldMonitorUnconfigured(t *testing.T) {
	t.Setenv("WORLDMONITOR_API_KEY", "")

	s := testServer(t)
	db := s.DB
	now := time.Now().UTC().Truncate(time.Millisecond)
	agent := dbpkg.Agent{
		ID:        uuid.NewString(),
		Name:      "wm-u-agent",
		APIKey:    "mb_wm_u_" + uuid.NewString(),
		CreatedAt: now,
	}
	if err := db.Create(&agent).Error; err != nil {
		t.Fatal(err)
	}
	q := dbpkg.FloorQuestion{
		ID:                   "Q-wm-u",
		Title:                "Q",
		Category:             "TEST",
		ResolutionCondition:  "x",
		Deadline:             "2026-12-31T00:00:00Z",
		Probability:          0.5,
		ProbabilityDelta:     0,
		AgentCount:           0,
		StakedCount:          0,
		Status:               "open",
		ClusterBreakdownJSON: `{}`,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := db.Create(&q).Error; err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/floor/questions/"+q.ID+"/context/worldmonitor", nil)
	req.Header.Set("Authorization", "Bearer "+agent.APIKey)
	res, err := ts.Client().Do(req)
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
	if body["live"] != false {
		t.Fatalf("expected live false: %#v", body["live"])
	}
}
