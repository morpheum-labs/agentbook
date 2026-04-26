package worldmon

import (
	"net/url"
	"strings"
	"testing"

	_ "embed"
)

// Embedded copy of the monitor-forge style library (e.g. from
// alohays/monitor-forge forge/data/rss-ls.json). `go test` checks every entry
// and every category without using the network.
//
//go:embed rss-ls.json
var rssLSEmbedded []byte

func TestRssLs_EveryEntry(t *testing.T) {
	f, err := ParseRSSLibraryJSON(rssLSEmbedded)
	if err != nil {
		t.Fatalf("parse rss-ls.json: %v", err)
	}
	if f.SchemaVersion == "" {
		t.Fatal("missing schemaVersion")
	}
	if n := len(f.Entries); n < 200 {
		t.Fatalf("vendored rss-ls.json: want at least 200 feed entries (monitor-forge scale), got %d", n)
	}
	for i, e := range f.Entries {
		if id := strings.TrimSpace(e.ID); id == "" {
			t.Errorf("entry %d: empty id", i)
		}
		if cat := strings.TrimSpace(e.Category); cat == "" {
			t.Errorf("entry %d (%s): empty category", i, e.ID)
		}
		u := strings.TrimSpace(e.URL)
		if u == "" {
			u = strings.TrimSpace(e.FallbackURL)
		}
		if u == "" {
			t.Errorf("entry %d (%s): no url and no fallbackUrl", i, e.ID)
			continue
		}
		pu, err := url.Parse(u)
		if err != nil {
			t.Errorf("entry %d (%s): bad url %q: %v", i, e.ID, u, err)
			continue
		}
		if pu.Scheme != "http" && pu.Scheme != "https" {
			t.Errorf("entry %d (%s): want http(s) url, got scheme %q", i, e.ID, pu.Scheme)
		}
	}
}

func TestRssLs_CategoriesResolveToURLs(t *testing.T) {
	f, err := ParseRSSLibraryJSON(rssLSEmbedded)
	if err != nil {
		t.Fatal(err)
	}
	byCat := make(map[string]int)
	seenID := make(map[string]int)
	for _, e := range f.Entries {
		byCat[strings.ToLower(strings.TrimSpace(e.Category))]++
		if e.ID != "" {
			seenID[e.ID]++
		}
	}
	// duplicate id check
	for id, n := range seenID {
		if n > 1 {
			t.Errorf("duplicate entry id: %q appears %d times", id, n)
		}
	}
	// every distinct category can drive at least one url via URLsForCategories
	for cat := range byCat {
		if cat == "" {
			t.Fatal("empty category key in map (should not happen)")
		}
		cat := cat
		t.Run("cat="+cat, func(t *testing.T) {
			t.Parallel()
			urls := f.URLsForCategories([]string{cat})
			if len(urls) < 1 {
				t.Fatalf("no urls for category %q", cat)
			}
		})
	}
	// Merged all feeds (de-duplicated) should be non-empty when the file has data.
	if all := f.URLsForCategories(nil); len(all) < 1 {
		t.Fatalf("URLsForCategories(nil): got 0, entries=%d distinctCats=%d", len(f.Entries), len(byCat))
	}
}
