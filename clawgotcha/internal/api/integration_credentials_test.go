//go:build integration

package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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

// TestIntegration_McpCredentialsReveal exercises instance secret + GET mcp-credentials.
func TestIntegration_McpCredentialsReveal(t *testing.T) {
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

	inst := "mcp-inst-" + strings.ReplaceAll(t.Name(), "/", "_")
	regBody := map[string]any{
		"instance_name": inst,
		"hostname":      "test",
		"version":       "1",
		"callback_url":  "http://127.0.0.1:9/cb",
	}
	rb, _ := json.Marshal(regBody)
	reqReg := httptest.NewRequest(http.MethodPost, "/api/v1/instances/register", bytes.NewReader(rb))
	reqReg.Header.Set("Content-Type", "application/json")
	recReg := httptest.NewRecorder()
	h.ServeHTTP(recReg, reqReg)
	if recReg.Code != http.StatusOK {
		t.Fatalf("register: %d %s", recReg.Code, recReg.Body.String())
	}
	var regWrap map[string]any
	if err := json.Unmarshal(recReg.Body.Bytes(), &regWrap); err != nil {
		t.Fatal(err)
	}
	sec, _ := regWrap["instance_api_secret"].(string)
	if sec == "" {
		t.Fatalf("expected instance_api_secret on first register: %v", regWrap)
	}

	agentName := "mcp-agent-" + strings.ReplaceAll(strings.ReplaceAll(t.Name(), "/", "_"), " ", "_")
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

	mcpName := "github-enterprise"
	createBody := map[string]any{
		"provider_slug":   "github",
		"label":           "default",
		"material_kind":   "github_pat",
		"mcp_server_name": mcpName,
		"plaintext":       "ghp_reveal_test_token",
	}
	cb, _ := json.Marshal(createBody)
	reqC := httptest.NewRequest(http.MethodPost, "/api/v1/agents/"+agentID+"/credentials", bytes.NewReader(cb))
	reqC.Header.Set("Content-Type", "application/json")
	recC := httptest.NewRecorder()
	h.ServeHTTP(recC, reqC)
	if recC.Code != http.StatusCreated {
		t.Fatalf("create credential: %d %s", recC.Code, recC.Body.String())
	}

	revealPath := "/api/v1/instances/" + url.PathEscape(inst) + "/agents/by-name/" + url.PathEscape(agentName) + "/mcp-credentials"
	reqBad := httptest.NewRequest(http.MethodGet, revealPath, nil)
	reqBad.Header.Set("X-Instance-Secret", "deadbeef")
	recBad := httptest.NewRecorder()
	h.ServeHTTP(recBad, reqBad)
	if recBad.Code != http.StatusForbidden {
		t.Fatalf("wrong secret want 403 got %d %s", recBad.Code, recBad.Body.String())
	}

	reqOk := httptest.NewRequest(http.MethodGet, revealPath, nil)
	reqOk.Header.Set("X-Instance-Secret", sec)
	recOk := httptest.NewRecorder()
	h.ServeHTTP(recOk, reqOk)
	if recOk.Code != http.StatusOK {
		t.Fatalf("reveal: %d %s", recOk.Code, recOk.Body.String())
	}
	var reveal map[string]any
	if err := json.Unmarshal(recOk.Body.Bytes(), &reveal); err != nil {
		t.Fatal(err)
	}
	bindings, _ := reveal["mcp_bindings"].([]any)
	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %v", reveal)
	}
	b0, _ := bindings[0].(map[string]any)
	if b0["mcp_server_name"] != mcpName {
		t.Fatalf("server name: %v", b0)
	}
}
