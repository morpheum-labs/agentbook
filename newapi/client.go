// Package newapi is a small Go client for [News API] (https://newsapi.org).
// The surface mirrors the Node "newsapi" client ([github.com/bzarras/newsapi]).
//
// [github.com/bzarras/newsapi]: https://github.com/bzarras/newsapi
// [News API]: https://newsapi.org
package newapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const defaultBaseURL = "https://newsapi.org"

// Client calls News API endpoints. Use [New] to create one.
type Client struct {
	apiKey   string
	baseURL  *url.URL
	corsPref string
	httpc    *http.Client

	V1 *V1
	V2 *V2
}

// New builds a [Client] with the given API key. The key is required (same as the
// Node client), and is sent as the X-Api-Key header on all requests that expect it
// except for [V1.Sources] (v1 /v1/sources), which matches the reference client and
// omits the key.
func New(apiKey string, opts ...ClientOption) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("newapi: no API key specified")
	}
	u, err := url.Parse(defaultBaseURL)
	if err != nil {
		return nil, err
	}
	c := &Client{
		apiKey:  apiKey,
		baseURL: u,
		httpc:   http.DefaultClient,
	}
	for _, o := range opts {
		if o == nil {
			continue
		}
		if err := o(c); err != nil {
			return nil, err
		}
	}
	c.V1 = &V1{c: c}
	c.V2 = &V2{c: c}
	return c, nil
}

// ClientOption configures [New] (CORS proxy URL, custom HTTP client, or base URL for tests).
type ClientOption func(*Client) error

// WithCORSProxyURL prepends a proxy origin to the request URL, as in
// the Node client's corsProxyUrl (e.g. "https://cors-anywhere.herokuapp.com/").
func WithCORSProxyURL(p string) ClientOption {
	return func(c *Client) error {
		c.corsPref = p
		return nil
	}
}

// WithHTTPClient uses a custom [http.Client] (timeouts, etc.).
func WithHTTPClient(h *http.Client) ClientOption {
	return func(c *Client) error {
		if h == nil {
			return fmt.Errorf("newapi: WithHTTPClient: nil *http.Client")
		}
		c.httpc = h
		return nil
	}
}

// WithBaseURL sets the host/scheme (default https://newsapi.org). Trailing slash is allowed.
func WithBaseURL(u string) ClientOption {
	return func(c *Client) error {
		p, err := url.Parse(u)
		if err != nil {
			return err
		}
		if p.Scheme == "" || p.Host == "" {
			return fmt.Errorf("newapi: WithBaseURL: need absolute URL with host")
		}
		c.baseURL = p
		return nil
	}
}

// RequestOption adjusts a single call (e.g. noCache, showHeaders from the reference client).
type RequestOption func(*reqOpts)

type reqOpts struct {
	noCache     bool
	showHeaders bool
}

// WithNoCache sets the X-No-Cache: true request header, matching the Node noCache: true
// per-request options.
func WithNoCache() RequestOption {
	return func(r *reqOpts) { r.noCache = true }
}

// WithShowHeaders makes API methods return non-nil [http.Response].Header in their second
// return value (when status is 200/JSON ok); otherwise the header return is nil.
func WithShowHeaders() RequestOption {
	return func(r *reqOpts) { r.showHeaders = true }
}

func collectRequestOptions(opts []RequestOption) reqOpts {
	var r reqOpts
	for _, o := range opts {
		if o == nil {
			continue
		}
		o(&r)
	}
	return r
}

func headOrNil(h http.Header, want bool) http.Header {
	if want {
		return h
	}
	return nil
}

// requestURL builds a URL the same way as the Node client: CORSProxyUrl + host + path + "?" + query
func (c *Client) requestURL(path string, q url.Values) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	base := strings.TrimRight(c.baseURL.String(), "/")
	s := c.corsPref + base + path
	if len(q) > 0 {
		s = s + "?" + q.Encode()
	}
	return s
}

func (c *Client) getJSON(ctx context.Context, urlStr string, withAPIKey bool, ro reqOpts) (http.Header, []byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, nil, err
	}
	if withAPIKey {
		req.Header.Set("X-Api-Key", c.apiKey)
	}
	// No-op in server-side Go; the Node client set these for browser fetches
	req.Header.Set("Access-Control-Allow-Origin", "*")
	if ro.noCache {
		req.Header.Set("X-No-Cache", "true")
	}
	res, err := c.httpc.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return res.Header, nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return res.Header, b, fmt.Errorf("newapi: HTTP %d", res.StatusCode)
	}
	// If JSON says status=error, surface [APIError]
	if err := checkJSONAPIError(b); err != nil {
		return res.Header, b, err
	}
	return res.Header, b, nil
}

// checkJSONAPIError inspects a News API body for {"status":"error",...}
func checkJSONAPIError(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	var w struct {
		Status  string `json:"status"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(b, &w); err != nil {
		return nil
	}
	if w.Status == "error" {
		return &APIError{Code: w.Code, Message: w.Message}
	}
	return nil
}

// APIError is returned when the response JSON has "status": "error".
type APIError struct {
	Code    string
	Message string
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" && e.Code != "" {
		return e.Code + ": " + e.Message
	}
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}

// V1 holds the legacy v1 API methods ([Client.V1]).
type V1 struct{ c *Client }

// V2 holds the v2 API methods ([Client.V2]).
type V2 struct{ c *Client }

func applyTopHeadlinesParams(p url.Values) url.Values {
	if p == nil {
		v := url.Values{}
		v.Set("language", "en")
		return v
	}
	return p
}
