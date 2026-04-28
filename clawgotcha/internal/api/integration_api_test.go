//go:build integration

package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/morpheumlabs/agentbook/clawgotcha/internal/api"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/db"
)

func TestIntegration_RegisterAndListInstances(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	g, err := db.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	h := api.NewRouter(g, api.RouterOptions{})
	body := map[string]any{
		"instance_name": "miroclaw-itest",
		"version":       "1.0.0",
		"hostname":      "test-host",
		"callback_url":  "http://127.0.0.1:9/webhook",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/instances/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("register: status %d body %s", rec.Code, rec.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/instances", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("list: status %d body %s", rec2.Code, rec2.Body.String())
	}
	var wrap map[string]any
	if err := json.Unmarshal(rec2.Body.Bytes(), &wrap); err != nil {
		t.Fatal(err)
	}
	arr, _ := wrap["instances"].([]any)
	if len(arr) < 1 {
		t.Fatalf("instances len: %v", len(arr))
	}
}
