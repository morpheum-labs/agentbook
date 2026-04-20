// Package worldmonitor calls World Monitor public HTTP APIs (see worldmonitor.app/docs).
// Agentglobe uses this only server-side; keys stay in WORLDMONITOR_API_KEY.
package worldmonitor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const defaultBaseURL = "https://worldmonitor.app"

// APIBase returns WORLDMONITOR_API_BASE or the production default (no trailing slash).
func APIBase() string {
	v := strings.TrimSpace(os.Getenv("WORLDMONITOR_API_BASE"))
	if v == "" {
		return defaultBaseURL
	}
	return strings.TrimRight(v, "/")
}

// APIKey returns WORLDMONITOR_API_KEY (X-WorldMonitor-Key header).
func APIKey() string {
	return strings.TrimSpace(os.Getenv("WORLDMONITOR_API_KEY"))
}

// Client is a thin HTTP wrapper for intelligence + forecast endpoints used as AgentFloor context.
type Client struct {
	BaseURL string
	Key     string
	HTTP    *http.Client
}

func NewClient() *Client {
	return &Client{
		BaseURL: APIBase(),
		Key:     APIKey(),
		HTTP: &http.Client{
			Timeout: 12 * time.Second,
		},
	}
}

func (c *Client) getJSON(ctx context.Context, path string, query url.Values) (json.RawMessage, int, error) {
	u, err := url.Parse(c.BaseURL + path)
	if err != nil {
		return nil, 0, err
	}
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, 0, err
	}
	if strings.TrimSpace(c.Key) != "" {
		req.Header.Set("X-WorldMonitor-Key", c.Key)
	}
	req.Header.Set("Accept", "application/json")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return nil, res.StatusCode, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return b, res.StatusCode, fmt.Errorf("worldmonitor: HTTP %d", res.StatusCode)
	}
	return json.RawMessage(b), res.StatusCode, nil
}

// FetchRiskScores calls GET /api/intelligence/v1/get-risk-scores.
func (c *Client) FetchRiskScores(ctx context.Context, region string) (json.RawMessage, error) {
	q := url.Values{}
	if strings.TrimSpace(region) != "" {
		q.Set("region", strings.TrimSpace(region))
	}
	body, code, err := c.getJSON(ctx, "/api/intelligence/v1/get-risk-scores", q)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("worldmonitor get-risk-scores: status %d body=%s", code, string(body))
	}
	return body, nil
}

// FetchForecasts calls GET /api/forecast/v1/get-forecasts.
func (c *Client) FetchForecasts(ctx context.Context, domain, region string) (json.RawMessage, error) {
	q := url.Values{}
	if strings.TrimSpace(domain) != "" {
		q.Set("domain", strings.TrimSpace(domain))
	}
	if strings.TrimSpace(region) != "" {
		q.Set("region", strings.TrimSpace(region))
	}
	body, code, err := c.getJSON(ctx, "/api/forecast/v1/get-forecasts", q)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("worldmonitor get-forecasts: status %d body=%s", code, string(body))
	}
	return body, nil
}
