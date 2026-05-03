package mcp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/morpheumlabs/agentbook/agentglobe/internal/config"
)

func TestEmbeddedMCPInitialize(t *testing.T) {
	cfg := &config.Config{
		Hostname:  "localhost:3456",
		Port:      3456,
		PublicURL: "http://127.0.0.1:3456",
	}
	h, err := EmbeddedMCPFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	body := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]any{
				"name":    "test",
				"version": "0.0.1",
			},
		},
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	var out struct {
		Result struct {
			ServerInfo struct {
				Name string `json:"name"`
			} `json:"serverInfo"`
		} `json:"result"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Result.ServerInfo.Name != "agentfloor" {
		t.Fatalf("server name: got %q", out.Result.ServerInfo.Name)
	}
}
