package worldmon

import (
	"context"
	"os"
	"testing"
	"time"
)

const miniLibrary = `{
  "schemaVersion": "1",
  "entries": [
    { "id": "a", "name": "A", "url": "https://example.com/a", "category": "politics" },
    { "id": "b", "name": "B", "url": "https://example.com/b", "category": "us" }
  ]
}`

func TestLoadRSSLibraryFile(t *testing.T) {
	_, err := LoadRSSLibraryFile("does-not-exist-xyz.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestParseRSSLibraryJSON_andURLsForCategories(t *testing.T) {
	f, err := ParseRSSLibraryJSON([]byte(miniLibrary))
	if err != nil {
		t.Fatal(err)
	}
	if f.SchemaVersion != "1" {
		t.Fatalf("schema: %q", f.SchemaVersion)
	}
	if len(f.Entries) != 2 {
		t.Fatalf("entries: %d", len(f.Entries))
	}
	one := f.URLsForCategories([]string{"politics"})
	if len(one) != 1 || one[0] != "https://example.com/a" {
		t.Fatalf("politics: %#v", one)
	}
	both := f.URLsForCategories([]string{"politics", "us"})
	if len(both) != 2 {
		t.Fatalf("two cats: %#v", both)
	}
	all := f.URLsForCategories(nil)
	if len(all) != 2 {
		t.Fatalf("all: %#v", all)
	}
	labeled := f.ForgeFeedsForCategories([]string{"politics", "us"})
	if len(labeled) != 2 {
		t.Fatalf("ForgeFeedsForCategories: got %#v", labeled)
	}
	if labeled[0].URL != "https://example.com/a" || labeled[0].Category != "politics" {
		t.Fatalf("first: %#v", labeled[0])
	}
	if f.ForgeFeedsForCategories(nil) != nil {
		t.Fatal(" ForgeFeedsForCategories(nil) want nil (no implied all-categories scan)")
	}
}

// Set MONITOR_FORGE_LIB_INTEGRATION=1 to assert the real GitHub JSON parses and
// returns URLs for common monitor-forge categories.
func TestIntegrationRSSLibraryFromMonitorForgeGitHub(t *testing.T) {
	if os.Getenv("MONITOR_FORGE_LIB_INTEGRATION") == "" {
		t.Skip("set MONITOR_FORGE_LIB_INTEGRATION=1 to verify live " + DefaultRSSLibraryURL)
	}
	t.Cleanup(resetRSSLibraryCacheForTest)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)
	// use uncached direct fetch
	lib, err := FetchRSSLibrary(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(lib.Entries) < 200 {
		t.Errorf("expected large curated library (200+), got %d", len(lib.Entries))
	}
	cases := []struct {
		cat string
		min int
	}{
		{"politics", 5},
		{"tech", 1},
		{"climate", 1},
	}
	for _, c := range cases {
		c := c
		t.Run(c.cat, func(t *testing.T) {
			t.Parallel()
			urls := lib.URLsForCategories([]string{c.cat})
			if len(urls) < c.min {
				t.Errorf("category %q: want at least %d urls, got %d", c.cat, c.min, len(urls))
			}
		})
	}
}
