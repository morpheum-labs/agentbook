package worldmon

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestIntelligenceGetRiskScores(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/intelligence/v1/get-risk-scores" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %q", r.Method)
		}
		if g := r.Header.Get(HeaderAPIKey); g != "k" {
			t.Errorf("key header = %q", g)
		}
		if r.URL.Query().Get("region") != "MENA" {
			t.Errorf("query region = %q", r.URL.Query().Get("region"))
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()
	cl := New("k", WithBaseURL(srv.URL))
	raw, err := cl.Intelligence().GetRiskScores(context.Background(), RiskScoresByRegion("MENA"))
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(raw) {
		t.Fatalf("invalid json: %s", raw)
	}
}

func TestServiceFetchV1Generic(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/maritime/v1/get-vessel-snapshot" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()
	cl := New("k", WithBaseURL(srv.URL))
	_, err := cl.Maritime().GetVesselSnapshot(context.Background(), url.Values{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestFetchV1(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	cl := New("", WithBaseURL(srv.URL))
	_, err := cl.FetchV1(context.Background(), "trade", "get-trade-barriers", nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestShippingV2Path(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/shipping/v2/list-webhooks" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()
	cl := New("x", WithBaseURL(srv.URL))
	_, err := cl.ShippingV2().ListWebhooks(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestServiceEmptyName(t *testing.T) {
	t.Parallel()
	cl := New("k")
	_, err := cl.Service("", "v1").Fetch(context.Background(), "m", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}
