package worldmon

import (
	"context"
	"encoding/json"
	"net/url"
)

// News is the local feed digest and optional upstream news helpers for /api/news/v1/…
type News struct{ *Service }

// News returns the news v1 service.
func (c *Client) News() *News { return &News{Service: c.Service("news", "v1")} }

// SummarizeArticle is GET /api/news/v1/summarize-article
func (n *News) SummarizeArticle(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return n.Fetch(ctx, "summarize-article", q)
}

// GetSummarizeArticleCache is GET /api/news/v1/get-summarize-article-cache
func (n *News) GetSummarizeArticleCache(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return n.Fetch(ctx, "get-summarize-article-cache", q)
}

// ListFeedDigest is local parallel RSS/Atom aggregation (see [ListFeedDigestLocal] and
// [AggregateFeeds]). The query may include:
//   - `feeds` (comma-separated RSS/Atom URLs), and/or
//   - `forge_categories` (comma-separated keys matching [RSSLibraryEntry.Category]
//     in [alohays/monitor-forge]’s [forge/data/rss-library.json] — fetched from
//     [DefaultRSSLibraryURL] unless [RSSLibraryEnv] is set;
//     optional `library_fresh` / `forge_fresh` / `forge_library_fresh` = 1|true to bypass
//     the in-process library cache, or set env MONITOR_FORGE_RSS_FRESH for the same);
//   - optional `limit` for max merged items; items from forge include `forgeCategory`
//     when the feed was resolved from `forge_categories`.
//
// [alohays/monitor-forge]: https://github.com/alohays/monitor-forge
// [forge/data/rss-library.json]: https://raw.githubusercontent.com/alohays/monitor-forge/main/forge/data/rss-library.json
func (n *News) ListFeedDigest(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return n.ListFeedDigestLocal(ctx, q)
}
