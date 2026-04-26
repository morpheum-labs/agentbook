package worldmon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// DefaultRSSLibraryURL is the published RSS feed library from
// [alohays/monitor-forge] (the same data the forge CLI’s `source list-library` uses).
//
// [alohays/monitor-forge]: https://github.com/alohays/monitor-forge
const DefaultRSSLibraryURL = "https://raw.githubusercontent.com/alohays/monitor-forge/main/forge/data/rss-library.json"

// RSSLibraryEnv overrides the library JSON URL (e.g. tests, air-gapped mirror).
const RSSLibraryEnv = "MONITOR_FORGE_RSS_LIBRARY_URL"

// RSSLibraryFile is the shape of monitor-forge’s forge/data/rss-library.json.
type RSSLibraryFile struct {
	SchemaVersion string            `json:"schemaVersion"`
	Entries       []RSSLibraryEntry `json:"entries"`
}

// RSSLibraryEntry is one curated feed row (see monitor-forge forge/data/rss-library.schema.ts).
type RSSLibraryEntry struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Category string `json:"category"`
	Language string `json:"language"`
	// If URL is empty, the CLI sometimes falls back to this.
	FallbackURL string `json:"fallbackUrl,omitempty"`
}

// ForgeFeed is a library entry after category filtering, used to tag digested
// [NewsItem] rows with the monitor-forge category that selected the feed.
type ForgeFeed struct {
	URL      string
	Category string
}

// LoadRSSLibraryFile reads a monitor-forge rss-library.json from a local path.
func LoadRSSLibraryFile(path string) (*RSSLibraryFile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseRSSLibraryJSON(b)
}

// ParseRSSLibraryJSON decodes a monitor-forge rss-library document.
func ParseRSSLibraryJSON(b []byte) (*RSSLibraryFile, error) {
	if len(b) == 0 {
		return nil, fmt.Errorf("empty rss library body")
	}
	var f RSSLibraryFile
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

// URLsForCategories returns each entry’s [RSSLibraryEntry.URL] when its
// category (case-insensitive) is listed in wantCat. If wantCat is empty, every
// non-empty url is included. Deduplication preserves first-seen order.
func (f *RSSLibraryFile) URLsForCategories(wantCat []string) []string {
	if f == nil {
		return nil
	}
	lower := make(map[string]struct{}, len(wantCat))
	for _, c := range wantCat {
		c = strings.ToLower(strings.TrimSpace(c))
		if c == "" {
			continue
		}
		lower[c] = struct{}{}
	}
	any := len(lower) == 0
	seen := make(map[string]struct{})
	var out []string
	for _, e := range f.Entries {
		cat := strings.ToLower(strings.TrimSpace(e.Category))
		if !any {
			if _, ok := lower[cat]; !ok {
				continue
			}
		}
		u := strings.TrimSpace(e.URL)
		if u == "" {
			u = strings.TrimSpace(e.FallbackURL)
		}
		if u == "" {
			continue
		}
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}

// ForgeFeedsForCategories is like [RSSLibraryFile.URLsForCategories] but returns
// the entry category for each non-duplicate feed URL. wantCat must be non-empty
// (otherwise an empty result is returned; "all categories" is not implied).
func (f *RSSLibraryFile) ForgeFeedsForCategories(wantCat []string) []ForgeFeed {
	if f == nil || len(wantCat) == 0 {
		return nil
	}
	lower := make(map[string]struct{}, len(wantCat))
	for _, c := range wantCat {
		c = strings.ToLower(strings.TrimSpace(c))
		if c == "" {
			continue
		}
		lower[c] = struct{}{}
	}
	if len(lower) == 0 {
		return nil
	}
	seen := make(map[string]struct{})
	var out []ForgeFeed
	for _, e := range f.Entries {
		cat := strings.ToLower(strings.TrimSpace(e.Category))
		if _, ok := lower[cat]; !ok {
			continue
		}
		u := strings.TrimSpace(e.URL)
		if u == "" {
			u = strings.TrimSpace(e.FallbackURL)
		}
		if u == "" {
			continue
		}
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, ForgeFeed{URL: u, Category: strings.TrimSpace(e.Category)})
	}
	return out
}

// envRSSLibraryAlwaysFresh, when "1" or "true" (see [isEnvTrue]), bypasses
// the in-process RSS library cache for [getCachedRSSLibrary] (same effect as
// a fresh request-scoped library fetch for forge category resolution).
const envRSSLibraryAlwaysFresh = "MONITOR_FORGE_RSS_FRESH"

// rssLibraryURL is the default or [RSSLibraryEnv] override.
func rssLibraryURL() string {
	if v := strings.TrimSpace(os.Getenv(RSSLibraryEnv)); v != "" {
		return strings.TrimSpace(v)
	}
	return DefaultRSSLibraryURL
}

var (
	libraryCacheMu  sync.Mutex
	libraryCached   *RSSLibraryFile
	libraryCachedAt time.Time
	libraryTTL      = 1 * time.Hour
)

// FetchRSSLibrary downloads and parses the current monitor-forge feed library
// (no in-memory cache). For repeated calls, use [getCachedRSSLibrary].
func FetchRSSLibrary(ctx context.Context, client *http.Client) (*RSSLibraryFile, error) {
	if client == nil {
		client = http.DefaultClient
	}
	u := rssLibraryURL()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json, */*;q=0.1")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, rerr := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if rerr != nil {
		return nil, rerr
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("rss library HTTP %d from %q", res.StatusCode, u)
	}
	return ParseRSSLibraryJSON(b)
}

// getCachedRSSLibrary fetches the monitor-forge library at most once per
// [libraryTTL]. If [envRSSLibraryAlwaysFresh] is "1" or "true", returns
// [FetchRSSLibrary] on every call without using the in-memory cache.
func getCachedRSSLibrary(ctx context.Context, client *http.Client) (*RSSLibraryFile, error) {
	if isEnvTrue(os.Getenv(envRSSLibraryAlwaysFresh)) {
		return FetchRSSLibrary(ctx, client)
	}
	libraryCacheMu.Lock()
	defer libraryCacheMu.Unlock()
	if libraryCached != nil && time.Since(libraryCachedAt) < libraryTTL {
		return libraryCached, nil
	}
	f, err := FetchRSSLibrary(ctx, client)
	if err != nil {
		return nil, err
	}
	libraryCached = f
	libraryCachedAt = time.Now()
	return f, nil
}

// forgeFeedsForCategories resolves monitor-forge category keys to
// [ForgeFeed] rows (by default the in-memory copy from [getCachedRSSLibrary]).
// If freshLibrary is true, [FetchRSSLibrary] is used so category→URL mappings
// match the latest published library. wantCat must contain at least one
// non-empty category.
func forgeFeedsForCategories(ctx context.Context, wantCat []string, freshLibrary bool) ([]ForgeFeed, error) {
	nonEmpty := 0
	for _, c := range wantCat {
		if strings.TrimSpace(c) != "" {
			nonEmpty++
		}
	}
	if nonEmpty == 0 {
		return nil, fmt.Errorf("forge categories: need at least one non-empty category key (comma-separated, e.g. tech,ai)")
	}
	var f *RSSLibraryFile
	var err error
	if freshLibrary {
		f, err = FetchRSSLibrary(ctx, nil)
	} else {
		f, err = getCachedRSSLibrary(ctx, nil)
	}
	if err != nil {
		return nil, err
	}
	out := f.ForgeFeedsForCategories(wantCat)
	if len(out) == 0 {
		return nil, fmt.Errorf("no feed URLs for categories %v in monitor-forge library", wantCat)
	}
	return out, nil
}

func isEnvTrue(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
