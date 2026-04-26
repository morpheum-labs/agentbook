package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/morpheumlabs/agentbook/worldmon"
)

// TestListFeedDigestHTTPLocal ensures GET /v1/wm/news/v1/list-feed-digest matches
// the path used by agentglobe MCP [get_world_context] for feed digests (cron / scheduled
// jobs that run without a third-party API base for news only).
func TestListFeedDigestHTTPLocal(t *testing.T) {
	t.Parallel()
	feed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<?xml version="1.0"?>
<rss version="2.0">
<channel><title>X</title>
  <item><title>From proxy test</title><link>http://x/1</link>
  <pubDate>Mon, 26 Apr 2026 10:00:00 GMT</pubDate></item>
</channel>
</rss>`))
	}))
	t.Cleanup(feed.Close)

	cl := worldmon.New("")

	r := NewRouter(cl, "http://127.0.0.1:9", "test", nil)
	uri := "/v1/wm/news/v1/list-feed-digest?feeds=" + url.QueryEscape(feed.URL)
	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodGet, uri, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", rr.Code, rr.Body.String())
	}
	var m struct {
		Count int              `json:"count"`
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&m); err != nil {
		t.Fatal(err)
	}
	if m.Count < 1 {
		t.Fatalf("count: %+v", m)
	}
	if m.Count != len(m.Items) {
		t.Errorf("count %d != len(items) %d", m.Count, len(m.Items))
	}
}

// TestListFeedDigestHTTPMissingFeeds returns 400 (matches strict ListFeedDigestLocal).
func TestListFeedDigestHTTPMissingFeeds(t *testing.T) {
	t.Parallel()
	rtr := NewRouter(worldmon.New(""), "http://127.0.0.1:9", "test", nil)
	rr := httptest.NewRecorder()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, "/v1/wm/news/v1/list-feed-digest", nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	rtr.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("code=%d want %d body=%s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}