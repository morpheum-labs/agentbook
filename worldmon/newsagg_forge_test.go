package worldmon

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// TestListFeedDigestLocal_ForgeCategories uses a local rss-library response and
// a tiny RSS feed; ensures monitor-forge category → URL → aggregate path works
// (same as ?forge_categories= on the worldmon HTTP proxy / MCP get_world_context).
func TestListFeedDigestLocal_ForgeCategories(t *testing.T) {
	t.Cleanup(resetRSSLibraryCacheForTest)
	rss := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<?xml version="1.0"?>
<rss version="2.0"><channel><title>T</title>
<item><title>Cat item</title><link>http://x/1</link>
<pubDate>Mon, 26 Apr 2026 10:00:00 GMT</pubDate></item>
</channel></rss>`))
	}))
	t.Cleanup(rss.Close)
	libBody := `{"schemaVersion":"1","entries":[` +
		`{"id":"m","name":"M","url":` + jsonString(rss.URL) + `,"category":"politics"}` +
		`]}`
	lib := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(libBody))
	}))
	t.Cleanup(lib.Close)
	t.Setenv(RSSLibraryEnv, lib.URL)

	n := New("").News()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)
	q := url.Values{
		"forge_categories": {"politics"},
		"limit":            {"10"},
	}
	raw, err := n.ListFeedDigestLocal(ctx, q)
	if err != nil {
		t.Fatal(err)
	}
	var m struct {
		Count int         `json:"count"`
		Items []NewsItem  `json:"items"`
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m.Count < 1 || len(m.Items) < 1 {
		t.Fatalf("count=%d items=%d", m.Count, len(m.Items))
	}
	if m.Items[0].ForgeCategory != "politics" {
		t.Errorf("got forgeCategory %q, want politics", m.Items[0].ForgeCategory)
	}
}

// jsonString returns a JSON string literal for s (quoted, escaped).
func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func TestListFeedDigestLocal_ForgeOnlyCommasRejects(t *testing.T) {
	n := New("").News()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	_, err := n.ListFeedDigestLocal(ctx, url.Values{"forge_categories": {",,"}})
	if err == nil {
		t.Fatal("expected error for no non-empty forge category tokens")
	}
}
