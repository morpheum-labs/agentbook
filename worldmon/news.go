package worldmon

import (
	"context"
	"encoding/json"
	"net/url"
)

// News is the worldmonitor/news/v1 API.
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

// ListFeedDigest is GET /api/news/v1/list-feed-digest
func (n *News) ListFeedDigest(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return n.Fetch(ctx, "list-feed-digest", q)
}
