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
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/credentials"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/db"
)

// TestIntegration_CredentialLifecycle exercises create → list → rotate → list and verifies no secret material in JSON.
func TestIntegration_CredentialLifecycle(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	g, err := db.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	mk, err := credentials.ParseMasterKey("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20")
	if err != nil {
		t.Fatal(err)
	}
	h := api.NewRouter(g, api.RouterOptions{CredentialsMasterKey: mk})

	agentName := "cred-test-" + t.Name()
	agentBody := map[string]any{
		"name":            agentName,
		"autonomy_level":  "ReadOnly",
		"system_prompt":   "x",
		"timeout_seconds": 30,
	}
	ab, _ := json.Marshal(agentBody)
	reqA := httptest.NewRequest(http.MethodPost, "/api/v1/agents", bytes.NewReader(ab))
	reqA.Header.Set("Content-Type", "application/json")
	recA := httptest.NewRecorder()
	h.ServeHTTP(recA, reqA)
	if recA.Code != http.StatusCreated {
		t.Fatalf("create agent: %d %s", recA.Code, recA.Body.String())
	}
	var agentWrap map[string]any
	if err := json.Unmarshal(recA.Body.Bytes(), &agentWrap); err != nil {
		t.Fatal(err)
	}
	ag, _ := agentWrap["agent"].(map[string]any)
	agentID, _ := ag["ID"].(string)
	if agentID == "" {
		t.Fatalf("missing agent ID: %v", agentWrap)
	}

	credPath := "/api/v1/agents/" + agentID + "/credentials"
	createBody := map[string]any{
		"provider_slug": "github",
		"label":         "default",
		"material_kind": "github_pat",
		"plaintext":     "ghp_test_not_real",
	}
	cb, _ := json.Marshal(createBody)
	reqC := httptest.NewRequest(http.MethodPost, credPath, bytes.NewReader(cb))
	reqC.Header.Set("Content-Type", "application/json")
	recC := httptest.NewRecorder()
	h.ServeHTTP(recC, reqC)
	if recC.Code != http.StatusCreated {
		t.Fatalf("create credential: %d %s", recC.Code, recC.Body.String())
	}
	var createWrap map[string]any
	if err := json.Unmarshal(recC.Body.Bytes(), &createWrap); err != nil {
		t.Fatal(err)
	}
	cr, _ := createWrap["credential"].(map[string]any)
	bindingID, _ := cr["id"].(string)
	if bindingID == "" {
		t.Fatalf("missing binding id: %v", createWrap)
	}
	if v, _ := cr["current_version"].(float64); int(v) != 1 {
		t.Fatalf("current_version want 1 got %v", cr["current_version"])
	}
	bodyStr := recC.Body.String()
	if strings.Contains(bodyStr, "ghp_test") || strings.Contains(bodyStr, "ciphertext") || strings.Contains(bodyStr, "nonce") {
		t.Fatalf("response leaked secret or raw crypto fields: %s", bodyStr)
	}

	reqL := httptest.NewRequest(http.MethodGet, credPath, nil)
	recL := httptest.NewRecorder()
	h.ServeHTTP(recL, reqL)
	if recL.Code != http.StatusOK {
		t.Fatalf("list: %d %s", recL.Code, recL.Body.String())
	}
	if strings.Contains(recL.Body.String(), "ghp_test") {
		t.Fatal("list leaked plaintext")
	}

	rotBody := map[string]any{"plaintext": "ghp_rotated_fake"}
	rb, _ := json.Marshal(rotBody)
	rotURL := credPath + "/" + bindingID + "/rotate"
	reqR := httptest.NewRequest(http.MethodPost, rotURL, bytes.NewReader(rb))
	reqR.Header.Set("Content-Type", "application/json")
	recR := httptest.NewRecorder()
	h.ServeHTTP(recR, reqR)
	if recR.Code != http.StatusOK {
		t.Fatalf("rotate: %d %s", recR.Code, recR.Body.String())
	}
	var rotWrap map[string]any
	if err := json.Unmarshal(recR.Body.Bytes(), &rotWrap); err != nil {
		t.Fatal(err)
	}
	c2, _ := rotWrap["credential"].(map[string]any)
	if v, _ := c2["current_version"].(float64); int(v) != 2 {
		t.Fatalf("after rotate current_version want 2 got %v", c2["current_version"])
	}
	if strings.Contains(recR.Body.String(), "ghp_rotated") {
		t.Fatal("rotate response leaked plaintext")
	}
}

func TestIntegration_CredentialCreateRequiresEncryptionKey(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	g, err := db.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	h := api.NewRouter(g, api.RouterOptions{})
	agentName := "cred-nokey-" + t.Name()
	agentBody := map[string]any{
		"name":            agentName,
		"autonomy_level":  "ReadOnly",
		"system_prompt":   "x",
		"timeout_seconds": 30,
	}
	ab, _ := json.Marshal(agentBody)
	reqA := httptest.NewRequest(http.MethodPost, "/api/v1/agents", bytes.NewReader(ab))
	reqA.Header.Set("Content-Type", "application/json")
	recA := httptest.NewRecorder()
	h.ServeHTTP(recA, reqA)
	if recA.Code != http.StatusCreated {
		t.Fatalf("create agent: %d %s", recA.Code, recA.Body.String())
	}
	var agentWrap map[string]any
	if err := json.Unmarshal(recA.Body.Bytes(), &agentWrap); err != nil {
		t.Fatal(err)
	}
	ag, _ := agentWrap["agent"].(map[string]any)
	agentID, _ := ag["ID"].(string)

	createBody := map[string]any{
		"provider_slug": "slack",
		"label":         "x",
		"material_kind": "api_key",
		"plaintext":     "x",
	}
	cb, _ := json.Marshal(createBody)
	reqC := httptest.NewRequest(http.MethodPost, "/api/v1/agents/"+agentID+"/credentials", bytes.NewReader(cb))
	reqC.Header.Set("Content-Type", "application/json")
	recC := httptest.NewRecorder()
	h.ServeHTTP(recC, reqC)
	if recC.Code != http.StatusServiceUnavailable {
		t.Fatalf("create without key: status %d body %s", recC.Code, recC.Body.String())
	}
}
