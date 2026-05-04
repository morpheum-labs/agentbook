package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublicURLFromClawgotchaCallback(t *testing.T) {
	tests := []struct {
		in     string
		want   string
		wantOK bool
	}{
		{"https://gw.example.com/webhook/clawgotcha", "https://gw.example.com", true},
		{"https://gw.example.com/webhook/clawgotcha/", "https://gw.example.com", true},
		{"https://gw.example.com/prefix/webhook/clawgotcha", "https://gw.example.com/prefix", true},
		{"http://127.0.0.1:18789/webhook/clawgotcha", "http://127.0.0.1:18789", true},
		{"https://gw.example.com/webhook/other", "", false},
		{"not-a-url", "", false},
		{"", "", false},
	}
	for _, tt := range tests {
		got, ok := publicURLFromClawgotchaCallback(tt.in)
		if ok != tt.wantOK || got != tt.want {
			t.Fatalf("publicURLFromClawgotchaCallback(%q) = (%q, %v); want (%q, %v)", tt.in, got, ok, tt.want, tt.wantOK)
		}
	}
}

func TestEffectiveInstancePublicURL(t *testing.T) {
	explicit := "https://explicit.example"
	got := effectiveInstancePublicURL(&explicit, "https://ignored/webhook/clawgotcha")
	if got == nil || *got != explicit {
		t.Fatalf("expected explicit public_url preserved")
	}
	empty := ""
	got2 := effectiveInstancePublicURL(&empty, "https://gw.example.com/webhook/clawgotcha")
	if got2 == nil || *got2 != "https://gw.example.com" {
		t.Fatalf("expected derive from callback, got %v", got2)
	}
	got3 := effectiveInstancePublicURL(nil, "https://gw.example.com/webhook/clawgotcha")
	if got3 == nil || *got3 != "https://gw.example.com" {
		t.Fatalf("expected derive when public_url nil")
	}
}

func TestMergeRegisterIngressMetadata(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/instances/register", bytes.NewReader(nil))
	req.Header.Set("X-Forwarded-Prefix", "/claw")
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "cp.example.com")
	raw := mergeRegisterIngressMetadata(json.RawMessage(`{"a":1}`), req)
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["a"].(float64) != 1 {
		t.Fatal("lost existing metadata")
	}
	ing, _ := m["clawgotcha_register_ingress"].(map[string]any)
	if ing == nil {
		t.Fatal("missing ingress")
	}
	if ing["request_path"] != "/api/v1/instances/register" {
		t.Fatalf("request_path: %v", ing["request_path"])
	}
	if ing["x_forwarded_prefix"] != "/claw" {
		t.Fatalf("prefix: %v", ing["x_forwarded_prefix"])
	}
}

func TestMergeRegisterIngressMetadata_noBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/prefix/api/v1/instances/register", nil)
	raw := mergeRegisterIngressMetadata(nil, req)
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	ing, _ := m["clawgotcha_register_ingress"].(map[string]any)
	if ing["request_path"] != "/prefix/api/v1/instances/register" {
		t.Fatal(ing)
	}
}
