// Command feed-digest resolves monitor-forge categories to feed URLs and writes JSON:
// either the merged digest (default) or only the URL list.
//
// Examples:
//
//	# Vendored library (from repo worldmon/):
//	go run ./cmd/feed-digest -library rss-ls.json -categories politics,us -out digest.json
//
//	# Fetch library from GitHub, print digest to stdout:
//	go run ./cmd/feed-digest -categories climate,tech -limit 50
//
//	# URL list only (no RSS fetch):
//	go run ./cmd/feed-digest -library rss-ls.json -categories ai -mode urls
//
// The HTTP / MCP digest endpoint supports ?library_fresh=1 (or ?forge_fresh=1) with
// ?forge_categories= so category→feed resolution uses the latest monitor-forge
// library JSON (bypassing the in-process 1h cache). Env MONITOR_FORGE_RSS_FRESH=1
// has the same effect. This CLI without -library already fetches a fresh library each run.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/morpheumlabs/agentbook/worldmon"
)

func main() {
	catStr := flag.String("categories", "", "comma-separated monitor-forge category keys (required), e.g. politics,us,tech")
	libraryPath := flag.String("library", "", "path to local rss-library.json; if empty, fetches from "+worldmon.DefaultRSSLibraryURL+" (or "+worldmon.RSSLibraryEnv+")")
	limit := flag.Int("limit", 100, "max items in the merged digest (digest mode only)")
	outPath := flag.String("out", "", "write JSON here; default stdout")
	mode := flag.String("mode", "digest", "digest (fetch RSS, merge) or urls (category → feed URL list only)")
	timeout := flag.Duration("timeout", 3*time.Minute, "HTTP timeout for library fetch and RSS fetches")
	flag.Parse()

	if strings.TrimSpace(*catStr) == "" {
		fmt.Fprintln(os.Stderr, "feed-digest: -categories is required")
		os.Exit(2)
	}
	cats := splitComma(*catStr)
	if len(cats) == 0 {
		fmt.Fprintln(os.Stderr, "feed-digest: no category values after parse")
		os.Exit(2)
	}

	var lib *worldmon.RSSLibraryFile
	if p := strings.TrimSpace(*libraryPath); p != "" {
		f, err := worldmon.LoadRSSLibraryFile(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "feed-digest: read library: %v\n", err)
			os.Exit(1)
		}
		lib = f
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		defer cancel()
		f, err := worldmon.FetchRSSLibrary(ctx, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "feed-digest: fetch library: %v\n", err)
			fmt.Fprintf(os.Stderr, "hint: pass -library path/to/rss-ls.json for offline use\n")
			os.Exit(1)
		}
		lib = f
	}

	feeds := lib.URLsForCategories(cats)
	if len(feeds) == 0 {
		fmt.Fprintf(os.Stderr, "feed-digest: no feed URLs for categories %v\n", cats)
		os.Exit(1)
	}

	m := strings.ToLower(strings.TrimSpace(*mode))
	switch m {
	case "urls":
		emitJSON(*outPath, map[string]any{
			"schemaVersion": lib.SchemaVersion,
			"categories":   cats,
			"feed_count":   len(feeds),
			"feed_urls":     feeds,
		})
		return
	case "digest":
		// break
	default:
		fmt.Fprintf(os.Stderr, "feed-digest: -mode must be digest or urls, got %q\n", *mode)
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	items, err := worldmon.AggregateFeeds(ctx, feeds, *limit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "feed-digest: %v\n", err)
		os.Exit(1)
	}
	emitJSON(*outPath, map[string]any{
		"categories": cats,
		"feed_count": len(feeds),
		"feed_urls":  feeds,
		"count":     len(items),
		"items":     items,
	})
}

func splitComma(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func emitJSON(path string, v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "feed-digest: json: %v\n", err)
		os.Exit(1)
	}
	if len(b) == 0 || b[len(b)-1] != '\n' {
		b = append(b, '\n')
	}
	if strings.TrimSpace(path) == "" {
		_, _ = os.Stdout.Write(b)
		return
	}
	if err := os.WriteFile(path, b, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "feed-digest: write %q: %v\n", path, err)
		os.Exit(1)
	}
}
