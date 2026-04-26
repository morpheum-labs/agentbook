// Package regclient posts capability registrations to agentglobe.
package regclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a thin HTTP client for the agentglobe service registry.
type Client struct {
	BaseURL   string
	Token     string
	UserAgent string
	HTTP      *http.Client
}

// NewClient builds a client. Base is the agentglobe origin, e.g. "https://globe.example.com".
// Token is SERVICE_REGISTRY_TOKEN.
func NewClient(baseURL, token, userAgent string) *Client {
	if userAgent == "" {
		userAgent = "morpheumlabs/newapi"
	}
	return &Client{
		BaseURL:   strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		Token:     strings.TrimSpace(token),
		UserAgent: userAgent,
		HTTP:      &http.Client{Timeout: 30 * time.Second},
	}
}

// Capable returns true if registration is configured.
func (c *Client) Capable() bool { return c != nil && c.BaseURL != "" && c.Token != "" }

// RegisterRequest is the body for POST .../capability-services/register.
type RegisterRequest struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	BaseURL      string   `json:"base_url"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	Tags         []string `json:"tags"`
	Domains      []string `json:"domains"`
	Metadata     any      `json:"metadata"`
	OpenapiURL   string   `json:"openapi_url"`
	OpenapiSpec  any      `json:"openapi_spec"`
	Status       string   `json:"status,omitempty"`
}

// Register calls agentglobe register. ctx may be request context; use context.Background for startup.
func (c *Client) Register(ctx context.Context, r RegisterRequest) error {
	if !c.Capable() {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if r.Tags == nil {
		r.Tags = []string{"news", "newapi.org"}
	}
	if r.Metadata == nil {
		r.Metadata = map[string]any{"kind": "newapi_server"}
	}
	return c.post(ctx, c.BaseURL+"/api/v1/capability-services/register", r)
}

// Heartbeat updates last_seen. Optional status: "active", "degraded", "inactive" (empty omits the field).
func (c *Client) Heartbeat(ctx context.Context, name, publicBase, status string) error {
	if !c.Capable() {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	body := map[string]any{
		"name":     name,
		"base_url": publicBase,
	}
	if strings.TrimSpace(status) != "" {
		body["status"] = strings.ToLower(strings.TrimSpace(status))
	}
	return c.post(ctx, c.BaseURL+"/api/v1/capability-services/heartbeat", body)
}

func (c *Client) post(ctx context.Context, rawURL string, body any) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("regclient: bad URL %q", rawURL)
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	hc := c.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	res, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("regclient: %s %d", u.Path, res.StatusCode)
	}
	return nil
}
