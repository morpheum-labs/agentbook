// Package worldmon is a Go client for the World Monitor HTTP APIs. Method paths
// follow the public server under server/worldmonitor on
// [World Monitor on GitHub] (e.g. GET /api/intelligence/v1/get-risk-scores).
// See [worldmonitor.app] for auth and product documentation.
//
// [World Monitor on GitHub]: https://github.com/koala73/worldmonitor/tree/main/server/worldmonitor
// [worldmonitor.app]: https://worldmonitor.app
package worldmon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const defaultBaseURL = "https://worldmonitor.app"

// Header for upstream authentication (set when you have a World Monitor key).
const HeaderAPIKey = "X-WorldMonitor-Key"

// DefaultWorldMonitorKeyEnv is the environment variable the upstream team documents for API keys.
const DefaultWorldMonitorKeyEnv = "WORLDMONITOR_API_KEY"

// DefaultWorldMonitorBaseEnv overrides the API base URL when set (e.g. WORLDMONITOR_API_BASE).
const DefaultWorldMonitorBaseEnv = "WORLDMONITOR_API_BASE"

// Client calls World Monitor JSON endpoints. It is safe for concurrent use
// (each request uses a snapshot of the configured HTTP client and base URL).
type Client struct {
	baseURL   string
	apiKey    string
	userAgent string
	http      *http.Client
}

// New builds a [Client] with the given API key (may be empty if a route does not
// require a key, though most production API paths expect one). Use [WithBaseURL] for
// staging, and [WithHTTPClient] to tune timeouts and tracing.
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		baseURL:   defaultBaseURL,
		apiKey:    strings.TrimSpace(apiKey),
		userAgent: "morpheumlabs/worldmon",
		http:      &http.Client{Timeout: 30 * time.Second},
	}
	for _, o := range opts {
		if o != nil {
			o(c)
		}
	}
	if c.baseURL == "" {
		c.baseURL = defaultBaseURL
	}
	if c.http == nil {
		c.http = &http.Client{Timeout: 30 * time.Second}
	}
	return c
}

// NewFromEnv is like [New] but takes the key from [DefaultWorldMonitorKeyEnv]
// and, when set, the base from [DefaultWorldMonitorBaseEnv].
func NewFromEnv(opts ...Option) *Client {
	key := strings.TrimSpace(os.Getenv(DefaultWorldMonitorKeyEnv))
	c := New(key, opts...)
	if b := strings.TrimSpace(os.Getenv(DefaultWorldMonitorBaseEnv)); b != "" {
		WithBaseURL(b)(c)
	}
	return c
}

// Option customizes [New].
type Option func(*Client)

// WithBaseURL sets the origin (e.g. https://worldmonitor.app) without a trailing path.
func WithBaseURL(u string) Option {
	return func(c *Client) {
		if s := strings.TrimSpace(u); s != "" {
			c.baseURL = strings.TrimRight(s, "/")
		}
	}
}

// WithHTTPClient sets the [http.Client] used for requests.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		if h != nil {
			c.http = h
		}
	}
}

// WithUserAgent sets the User-Agent header. Empty leaves the default.
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		if s := strings.TrimSpace(ua); s != "" {
			c.userAgent = s
		}
	}
}

// APIKey returns the key configured for this client (read-only; set only via [New]).
func (c *Client) APIKey() string { return c.apiKey }

// BaseURL returns the configured base URL, without trailing slash.
func (c *Client) BaseURL() string { return c.baseURL }

// Service returns a [Service] for arbitrary /api/{name}/{version}/… paths.
// Use the typed methods on [Client] when you want a narrower surface.
func (c *Client) Service(name, version string) *Service {
	if c == nil {
		return nil
	}
	if strings.TrimSpace(name) == "" {
		return &Service{client: c, name: "", version: version}
	}
	if version == "" {
		version = "v1"
	}
	return &Service{client: c, name: strings.Trim(name, "/"), version: version}
}

// FetchV1 is shorthand for s.Fetch with version "v1" at the first path component service.
// Example: FetchV1(ctx, "maritime", "get-vessel-snapshot", q) -> GET /api/maritime/v1/get-vessel-snapshot
func (c *Client) FetchV1(ctx context.Context, service, method string, q url.Values) (json.RawMessage, error) {
	return c.Service(service, "v1").Fetch(ctx, method, q)
}

// FetchV2 calls /api/{service}/v2/{method}.
func (c *Client) FetchV2(ctx context.Context, service, method string, q url.Values) (json.RawMessage, error) {
	return c.Service(service, "v2").Fetch(ctx, method, q)
}

func (c *Client) endpointURL(path string, q url.Values) string {
	base := strings.TrimRight(c.baseURL, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	s := base + path
	if q != nil && len(q) > 0 {
		if enc := q.Encode(); enc != "" {
			s += "?" + enc
		}
	}
	return s
}

func (c *Client) doGet(ctx context.Context, fullPath string) (json.RawMessage, int, error) {
	if c == nil {
		return nil, 0, errors.New("worldmon: nil client")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullPath, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Accept", "application/json")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	if c.apiKey != "" {
		req.Header.Set(HeaderAPIKey, c.apiKey)
	}
	res, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return b, res.StatusCode, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		trim := string(b)
		if len(trim) > 512 {
			trim = trim[:512] + "…"
		}
		return b, res.StatusCode, fmt.Errorf("worldmon: HTTP %d: %s", res.StatusCode, trim)
	}
	return json.RawMessage(b), res.StatusCode, nil
}
