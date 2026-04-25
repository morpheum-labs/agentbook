package worldmon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// Service is a versioned World Monitor service (maps to /api/{name}/{version}/…).
// Obtain one with [Client.Service] or the helpers on [Client] (e.g. [Client.Intelligence]).
type Service struct {
	client  *Client
	name    string
	version string
}

// Fetch issues GET /api/{service}/{version}/{method} with optional query.
// The method string is the final path segment, usually kebab-case and matching the
// handler file name in server/worldmonitor/... (without .ts).
func (s *Service) Fetch(ctx context.Context, method string, q url.Values) (json.RawMessage, error) {
	if s == nil || s.client == nil {
		return nil, errors.New("worldmon: nil Service or Client")
	}
	if strings.TrimSpace(s.name) == "" {
		return nil, errors.New("worldmon: empty service name; use Client.Service or a Client.*() helper")
	}
	m := strings.Trim(strings.TrimSpace(method), "/")
	if m == "" {
		return nil, errors.New("worldmon: empty method path")
	}
	v := s.version
	if v == "" {
		v = "v1"
	}
	path := fmt.Sprintf("/api/%s/%s/%s", s.name, v, m)
	u := s.client.endpointURL(path, q)
	return s.fetchURL(ctx, u)
}

// Name and Version return the first two path segments after /api/ (for tests and logging).
func (s *Service) Name() string { return s.name }

// Version returns the version segment, e.g. "v1" or "v2".
func (s *Service) Version() string { return s.version }

func (s *Service) fetchURL(ctx context.Context, full string) (json.RawMessage, error) {
	b, _, err := s.client.doGet(ctx, full)
	return b, err
}
