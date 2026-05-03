package worldmon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
)

// NewsItem is a normalized item returned by the local RSS/Atom aggregator.
type NewsItem struct {
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	GUID        string    `json:"guid,omitempty"`
	PubDate     time.Time `json:"pubDate"`
	Description string    `json:"description,omitempty"`
	Author      string    `json:"author,omitempty"`
	Thumbnail   string    `json:"thumbnail,omitempty"`
	Enclosure   string    `json:"enclosure,omitempty"`
	FeedTitle   string    `json:"feedTitle,omitempty"`
	FeedURL     string    `json:"feedURL,omitempty"`
	// ForgeCategory is the monitor-forge [RSSLibraryEntry.Category] for the feed
	// that produced this item (when the digest used ?forge_categories= or equivalent).
	ForgeCategory string `json:"forgeCategory,omitempty"`
}

// AggregateFeeds fetches multiple RSS/Atom feeds in parallel, normalizes, sorts
// newest-first, and returns items (truncated to the given limit).
func AggregateFeeds(ctx context.Context, feedURLs []string, limit int) ([]NewsItem, error) {
	return aggregateFeedsWithForgeMap(ctx, feedURLs, nil, limit)
}

func aggregateFeedsWithForgeMap(ctx context.Context, feedURLs []string, forgeByFeedURL map[string]string, limit int) ([]NewsItem, error) {
	if len(feedURLs) == 0 {
		return nil, fmt.Errorf("no feeds provided")
	}
	if limit <= 0 {
		limit = 100
	}

	fp := gofeed.NewParser()
	var wg sync.WaitGroup
	itemsCh := make(chan []NewsItem, len(feedURLs))

	for _, u := range feedURLs {
		wg.Add(1)
		feedURL := u
		forgeCat := ""
		if forgeByFeedURL != nil {
			forgeCat = forgeByFeedURL[feedURL]
		}
		go func() {
			defer wg.Done()
			feed, err := fp.ParseURLWithContext(feedURL, ctx)
			if err != nil {
				// per-feed errors are ignored (tolerant merge)
				return
			}

			var batch []NewsItem
			for _, item := range feed.Items {
				pubDate := item.PublishedParsed
				if pubDate == nil {
					pubDate = item.UpdatedParsed
				}
				if pubDate == nil {
					continue
				}

				batch = append(batch, NewsItem{
					Title:         item.Title,
					Link:          item.Link,
					GUID:          item.GUID,
					PubDate:       *pubDate,
					Description:   item.Description,
					Author:        itemAuthor(item),
					Thumbnail:     getThumbnail(item),
					Enclosure:     getEnclosure(item),
					FeedTitle:     feed.Title,
					FeedURL:       feedURL,
					ForgeCategory: forgeCat,
				})
			}
			itemsCh <- batch
		}()
	}

	go func() {
		wg.Wait()
		close(itemsCh)
	}()

	var all []NewsItem
	for batch := range itemsCh {
		all = append(all, batch...)
	}

	// sort newest first
	sort.Slice(all, func(i, j int) bool {
		return all[i].PubDate.After(all[j].PubDate)
	})

	if len(all) > limit {
		all = all[:limit]
	}

	return all, nil
}

func itemAuthor(item *gofeed.Item) string {
	if item == nil {
		return ""
	}
	if item.Author != nil && item.Author.Name != "" {
		return item.Author.Name
	}
	if len(item.Authors) > 0 && item.Authors[0] != nil {
		return item.Authors[0].Name
	}
	return ""
}

func getThumbnail(item *gofeed.Item) string {
	if item == nil {
		return ""
	}
	if item.Image != nil && item.Image.URL != "" {
		return item.Image.URL
	}
	return ""
}

func getEnclosure(item *gofeed.Item) string {
	if item == nil {
		return ""
	}
	for _, e := range item.Enclosures {
		if e != nil && e.URL != "" {
			return e.URL
		}
	}
	return ""
}

// ListFeedDigestLocal builds a `{"items", "count"}` JSON object. Supply either
//   - `feeds` — comma-separated feed URLs, and/or
//   - `forge_categories` — comma-separated monitor-forge [RSSLibraryEntry.Category]
//     values (e.g. politics,us,tech); feed URLs are resolved from
//     [DefaultRSSLibraryURL] (override with [RSSLibraryEnv]).
//
// Optional: `library_fresh`, `forge_fresh`, or `forge_library_fresh` = 1|true|yes|on
// fetches a fresh [RSSLibraryFile] (no in-process cache) before resolving those categories
// to feed URLs, so the library matches the latest published monitor-forge data.
//
// Optional: `rss_library` — absolute path to a local rss-library.json on the worldmon host;
// `rss_library_url` — HTTPS/HTTP URL to a library (per-request fetch). If both are set,
// `rss_library` wins. (af-local-mcp can supply them from the rss_lib config field.)
// Both `feeds` and `forge_categories` may be set; URLs are merged and de-duplicated
// (first occurrence wins: explicit `feeds` URLs are listed before forge resolution).
func (n *News) ListFeedDigestLocal(ctx context.Context, q url.Values) (json.RawMessage, error) {
	feedsParam := q.Get("feeds")
	forgeCat := q.Get("forge_categories")
	if feedsParam == "" && forgeCat == "" {
		return nil, fmt.Errorf("missing ?feeds= and/or ?forge_categories= (comma-separated values); forge_categories use keys from alohays/monitor-forge forge/data/rss-library.json")
	}
	var expl []string
	if feedsParam != "" {
		expl = append(expl, splitAndTrim(feedsParam, ",")...)
	}
	var refs []ForgeFeed
	if forgeCat != "" {
		cats := splitAndTrim(forgeCat, ",")
		if len(cats) == 0 {
			return nil, fmt.Errorf("forge_categories must name at least one non-empty category (e.g. tech or tech,ai; got only commas or space)")
		}
		fresh := queryParamTrue(q, "library_fresh", "forge_fresh", "forge_library_fresh")
		var err error
		refs, err = forgeFeedsForCategories(ctx, cats, fresh, q.Get("rss_library"), q.Get("rss_library_url"))
		if err != nil {
			return nil, err
		}
	}

	// Pre-dedup combined order: explicit feed URLs first, then forge, so
	// a duplicate URL prefers explicit and gets no `forgeCategory` on items.
	var combined []string
	for _, u := range expl {
		if t := strings.TrimSpace(u); t != "" {
			combined = append(combined, t)
		}
	}
	nExplicit := len(combined)
	for _, r := range refs {
		if t := strings.TrimSpace(r.URL); t != "" {
			combined = append(combined, t)
		}
	}
	feedURLs := dedupeURLOrdered(combined)
	if len(feedURLs) == 0 {
		return nil, fmt.Errorf("no feed URLs after resolving feeds and forge_categories")
	}

	var forgeByFeedURL map[string]string
	if len(refs) > 0 {
		firstIdx := make(map[string]int, len(combined))
		for i, u := range combined {
			u = strings.TrimSpace(u)
			if u == "" {
				continue
			}
			if _, ok := firstIdx[u]; !ok {
				firstIdx[u] = i
			}
		}
		refCat := make(map[string]string, len(refs))
		for _, r := range refs {
			if u := strings.TrimSpace(r.URL); u != "" {
				if _, ok := refCat[u]; !ok {
					refCat[u] = r.Category
				}
			}
		}
		forgeByFeedURL = make(map[string]string, len(refCat))
		for u := range refCat {
			if firstIdx[u] >= nExplicit {
				forgeByFeedURL[u] = refCat[u]
			}
		}
	}

	limit := 100
	if s := q.Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			limit = n
		}
	}
	items, err := aggregateFeedsWithForgeMap(ctx, feedURLs, forgeByFeedURL, limit)
	if err != nil {
		return nil, err
	}

	resp := map[string]any{
		"items": items,
		"count": len(items),
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

func queryParamTrue(q url.Values, keys ...string) bool {
	for _, k := range keys {
		if t := strings.ToLower(strings.TrimSpace(q.Get(k))); t == "1" || t == "true" || t == "yes" || t == "on" {
			return true
		}
	}
	return false
}

func splitAndTrim(s, sep string) []string {
	var res []string
	for _, v := range strings.Split(s, sep) {
		if t := strings.TrimSpace(v); t != "" {
			res = append(res, t)
		}
	}
	return res
}

func dedupeURLOrdered(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	var out []string
	for _, u := range in {
		u = strings.TrimSpace(u)
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
