package httpapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestHandlePublicWorldContext_ProxiesToWorldmon(t *testing.T) {
	var gotPath, gotQuery string
	wm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"ok":true,"digest":[]}`)
	}))
	t.Cleanup(wm.Close)
	t.Setenv("WORLDMON_BASE_URL", wm.URL)
	t.Setenv("CONFIG_PATH", "")

	s := testServer(t)
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)

	u, err := url.Parse(ts.URL + "/api/v1/public/world-context")
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("method", "list-feed-digest")
	q.Set("forge_categories", "politics")
	u.RawQuery = q.Encode()
	res, err := http.Get(u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { res.Body.Close() })
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d: %s", res.StatusCode, string(b))
	}
	if gotPath != "/v1/wm/news/v1/list-feed-digest" {
		t.Fatalf("path: %q", gotPath)
	}
	if !strings.Contains(gotQuery, "forge_categories=politics") {
		t.Fatalf("query: %q", gotQuery)
	}
}

func TestHandlePublicWorldContext_MethodParamCaseInsensitive(t *testing.T) {
	wm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{}`)
	}))
	t.Cleanup(wm.Close)
	t.Setenv("WORLDMON_BASE_URL", wm.URL)

	s := testServer(t)
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)

	res, err := http.Get(ts.URL + "/api/v1/public/world-context?Method=list-feed-digest&feeds=*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { res.Body.Close() })
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d: %s", res.StatusCode, string(b))
	}
}

func TestHandlePublicWorldContext_RequiresMethod(t *testing.T) {
	s := testServer(t)
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	res, err := http.Get(ts.URL + "/api/v1/public/world-context")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { res.Body.Close() })
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d", res.StatusCode)
	}
}
