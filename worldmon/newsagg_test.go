package worldmon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
  <title>Test Feed</title>
  <item>
    <title>Scheduled test headline</title>
    <link>http://127.0.0.1/item/1</link>
    <pubDate>Mon, 02 Jan 2006 15:04:05 GMT</pubDate>
  </item>
  <item>
    <title>Second item</title>
    <link>http://127.0.0.1/item/2</link>
    <pubDate>Mon, 26 Apr 2026 12:00:00 GMT</pubDate>
  </item>
</channel>
</rss>`

func TestAggregateFeeds_httptestRSS(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(testRSS))
	}))
	t.Cleanup(srv.Close)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	items, err := AggregateFeeds(ctx, []string{srv.URL}, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) < 2 {
		t.Fatalf("expected 2+ items, got %d: %+v", len(items), items)
	}
	if got := items[0].Title; got != "Second item" {
		// 2026 sorts after 2006 — newest first
		t.Errorf("newest first: want title %q, got %q", "Second item", got)
	}
}

func TestStringFromEnv_PrefersNewNames(t *testing.T) {
	t.Setenv(EnvAPIKey, "newv")
	t.Setenv(EnvAPIKeyLegacy, "oldv")
	if s := StringFromEnv(EnvAPIKey, EnvAPIKeyLegacy); s != "newv" {
		t.Errorf("StringFromEnv = %q", s)
	}
}
