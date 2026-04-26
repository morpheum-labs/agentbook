package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	mcpg "github.com/metoro-io/mcp-golang"
)

// DefaultDailyNewsAPIBase is the public API used by the daily-news MCP (6551 open endpoints).
const DefaultDailyNewsAPIBase = "https://ai.6551.io"

// GetHotNewsArgs matches daily-news "get_hot_news" and REST /open/free_hot.
type GetHotNewsArgs struct {
	Category     string `json:"category" jsonschema:"required,description=Top-level news category key (e.g. crypto, ai, business)"`
	Subcategory  string `json:"subcategory" jsonschema:"description=Optional subcategory key (e.g. defi) when the category has subcategories"`
	Limit        int    `json:"limit" jsonschema:"description=Optional max rows hint (passed when the upstream API supports it)"`
	Timeframe    string `json:"timeframe" jsonschema:"description=Optional time window label for the client; may be ignored by the upstream open API"`
}

// GetNewsCategoriesArgs is an empty argument set (categories list is global).
type GetNewsCategoriesArgs struct{}

func (s *State) getHotNews(ctx context.Context, args GetHotNewsArgs) (*mcpg.ToolResponse, error) {
	_ = ctx
	if strings.TrimSpace(args.Category) == "" {
		return nil, fmt.Errorf("category is required")
	}
	base := strings.TrimRight(s.DailyNewsAPIBase, "/")
	if base == "" {
		base = DefaultDailyNewsAPIBase
	}
	u, err := url.Parse(base + "/open/free_hot")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("category", strings.TrimSpace(args.Category))
	if sc := strings.TrimSpace(args.Subcategory); sc != "" {
		q.Set("subcategory", sc)
	}
	if args.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", args.Limit))
	}
	u.RawQuery = q.Encode()

	body, err := s.getHTTP(ctx, u.String())
	if err != nil {
		return nil, err
	}
	return toolTextJSON(string(body))
}

func (s *State) getNewsCategories(ctx context.Context, _ GetNewsCategoriesArgs) (*mcpg.ToolResponse, error) {
	base := strings.TrimRight(s.DailyNewsAPIBase, "/")
	if base == "" {
		base = DefaultDailyNewsAPIBase
	}
	u := base + "/open/free_categories"
	body, err := s.getHTTP(ctx, u)
	if err != nil {
		return nil, err
	}
	return toolTextJSON(string(body))
}

func (s *State) getHTTP(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", s.userAgentHeader())
	res, err := s.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, rerr := io.ReadAll(res.Body)
	if rerr != nil {
		return nil, rerr
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("http %d: %s", res.StatusCode, strings.TrimSpace(string(b)))
	}
	return b, nil
}

func toolTextJSON(s string) (*mcpg.ToolResponse, error) {
	// Re-indent so clients see valid JSON
	var v any
	if err := json.Unmarshal([]byte(s), &v); err == nil {
		pretty, err2 := json.Marshal(v)
		if err2 == nil {
			return mcpg.NewToolResponse(mcpg.NewTextContent(string(pretty))), nil
		}
	}
	return mcpg.NewToolResponse(mcpg.NewTextContent(s)), nil
}

func (s *State) userAgentHeader() string {
	if s != nil && s.UserAgent != "" {
		return s.UserAgent
	}
	return "agentglobe-mcp/1.0"
}

func (s *State) httpClient() *http.Client {
	if s == nil || s.HTTPClient == nil {
		return defaultHTTPClient()
	}
	return s.HTTPClient
}

// Short timeout for outbound calls if the default client is used.
func defaultHTTPClient() *http.Client {
	return &http.Client{Timeout: 60 * time.Second}
}
