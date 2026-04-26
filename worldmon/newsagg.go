package worldmon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
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
}

// AggregateFeeds fetches multiple RSS/Atom feeds in parallel, normalizes, sorts
// newest-first, and returns items (truncated to the given limit).
func AggregateFeeds(ctx context.Context, feedURLs []string, limit int) ([]NewsItem, error) {
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
		go func(feedURL string) {
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
					Title:       item.Title,
					Link:        item.Link,
					GUID:        item.GUID,
					PubDate:     *pubDate,
					Description: item.Description,
					Author:      itemAuthor(item),
					Thumbnail:   getThumbnail(item),
					Enclosure:   getEnclosure(item),
					FeedTitle:   feed.Title,
					FeedURL:     feedURL,
				})
			}
			itemsCh <- batch
		}(u)
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

// ListFeedDigestLocal builds a `{"items", "count"}` JSON object from the `feeds`
// query param (comma-separated feed URLs).
func (n *News) ListFeedDigestLocal(ctx context.Context, q url.Values) (json.RawMessage, error) {
	feedsParam := q.Get("feeds")
	if feedsParam == "" {
		return nil, fmt.Errorf("missing ?feeds= parameter (comma-separated URLs)")
	}

	feedURLs := splitAndTrim(feedsParam, ",")
	items, err := AggregateFeeds(ctx, feedURLs, 100)
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

func splitAndTrim(s, sep string) []string {
	var res []string
	for _, v := range strings.Split(s, sep) {
		if t := strings.TrimSpace(v); t != "" {
			res = append(res, t)
		}
	}
	return res
}
