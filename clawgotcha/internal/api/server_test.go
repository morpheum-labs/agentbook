package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAPISpec(t *testing.T) {
	b, err := readEmbeddedOpenapi()
	if err != nil {
		t.Fatalf("readEmbeddedOpenapi: %v", err)
	}
	var spec map[string]any
	if err := json.Unmarshal(b, &spec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if spec["openapi"] != "3.0.3" {
		t.Fatalf("unexpected openapi: %v", spec["openapi"])
	}
}

func TestOpenAPIRoute(t *testing.T) {
	r := NewRouter(nil, RouterOptions{})
	req := httptest.NewRequest("GET", "/openapi.json", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatalf("content-type: %q", rec.Header().Get("Content-Type"))
	}
}

func TestCORS_PreflightEchoRequestHeaders(t *testing.T) {
	r := NewRouter(nil, RouterOptions{})
	req := httptest.NewRequest("OPTIONS", "/api/v1/agents", nil)
	req.Header.Set("Origin", "https://app.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "accept,content-type,authorization")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("preflight status: got %d want %d", rec.Code, http.StatusNoContent)
	}
	if g, w := rec.Header().Get("Access-Control-Allow-Origin"), "*"; g != w {
		t.Fatalf("Allow-Origin: got %q want %q", g, w)
	}
	if g, w := rec.Header().Get("Access-Control-Allow-Headers"), "accept,content-type,authorization"; g != w {
		t.Fatalf("Allow-Headers echo: got %q want %q", g, w)
	}
	if g := rec.Header().Get("Access-Control-Allow-Methods"); g == "" {
		t.Fatal("expected Access-Control-Allow-Methods on preflight")
	}
}

func TestCORS_GETWithOrigin(t *testing.T) {
	r := NewRouter(nil, RouterOptions{})
	req := httptest.NewRequest("GET", "/healthz", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if g, w := rec.Header().Get("Access-Control-Allow-Origin"), "*"; g != w {
		t.Fatalf("Allow-Origin: got %q want %q", g, w)
	}
}

func TestMetricsRoute(t *testing.T) {
	r := NewRouter(nil, RouterOptions{})
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Fatalf("content-type: %q", ct)
	}
}

func TestAPIKey_BlocksAPIWithoutKey(t *testing.T) {
	r := NewRouter(nil, RouterOptions{APIKey: "secret"})
	req := httptest.NewRequest("GET", "/api/v1/agents", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestAPIKey_PublicRoutesWithoutKey(t *testing.T) {
	r := NewRouter(nil, RouterOptions{APIKey: "secret"})
	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("healthz status %d", rec.Code)
	}
}
