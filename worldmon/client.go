// Package worldmon is a Go HTTP client for versioned JSON APIs of the form
// GET /api/{service}/{version}/{method} with query parameters, optional
// [HeaderAPIKey] and [StringFromEnv] configuration.
// Error bodies are parsed with [ParseErrorBody]; [CacheTierForPath] and related
// helpers follow common “gateway cache tier” conventions.
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

const defaultBaseURL = ""

// Header for upstream API authentication when the key is set (X-API-Key style; some
// gateways also accept [HeaderAPIKeyAlt] or other product-specific key headers).
const HeaderAPIKey = "X-API-Key"

// HeaderAPIKeyAlt is a common alternate header name.
const HeaderAPIKeyAlt = "X-Api-Key"

// Environment variable names.
const (
	EnvAPIKey  = "WORLDMON_API_KEY"
	EnvBaseURL = "WORLDMON_API_BASE"
	// EnvAPIKeyLegacy and EnvBaseURLLegacy are deprecated but still read by [StringFromEnv] and [NewFromEnv] when
	// the preferred [EnvAPIKey] / [EnvBaseURL] are empty.
	EnvAPIKeyLegacy  = "WORLDMONITOR_API_KEY"
	EnvBaseURLLegacy = "WORLDMONITOR_API_BASE"
)

// StringFromEnv returns the first non-empty trimmed value of os.Getenv among keys.
func StringFromEnv(keys ...string) string {
	for _, k := range keys {
		if s := strings.TrimSpace(os.Getenv(k)); s != "" {
			return s
		}
	}
	return ""
}

// Client is an HTTP client for a configured API origin. It is safe for concurrent
// use (each request uses a snapshot of the configured [http.Client] and base URL).
type Client struct {
	baseURL   string
	apiKey    string
	userAgent string
	http      *http.Client
}

// New builds a [Client] with the given API key (optional). Use [WithBaseURL] to
// set the request origin. When the base is empty, remote [Service] calls need
// [WithBaseURL] or [StringFromEnv]([EnvBaseURL], [EnvBaseURLLegacy]). Use [WithHTTPClient] to
// tune timeouts and tracing.
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
	// do not backfill a hard-coded host; the caller or [WithBaseURL] / env must set it
	if c.http == nil {
		c.http = &http.Client{Timeout: 30 * time.Second}
	}
	return c
}

// NewFromEnv is like [New] with the key and base from [StringFromEnv]([EnvAPIKey], [EnvAPIKeyLegacy]) and
// [StringFromEnv]([EnvBaseURL], [EnvBaseURLLegacy]).
func NewFromEnv(opts ...Option) *Client {
	key := StringFromEnv(EnvAPIKey, EnvAPIKeyLegacy)
	c := New(key, opts...)
	if b := StringFromEnv(EnvBaseURL, EnvBaseURLLegacy); b != "" {
		WithBaseURL(b)(c)
	}
	return c
}

// Option customizes [New].
type Option func(*Client)

// WithBaseURL sets the origin (scheme+host) without a trailing path.
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
