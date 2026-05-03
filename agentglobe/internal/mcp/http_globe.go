package mcp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (s *State) globeURL(apiPath string) string {
	p := strings.TrimSpace(apiPath)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return strings.TrimRight(strings.TrimSpace(s.GlobeBaseURL), "/") + p
}

func (s *State) httpDo(ctx context.Context, method, apiPath string, h http.Header, body []byte) ([]byte, int, error) {
	var rdr io.Reader
	if len(body) > 0 {
		rdr = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, s.globeURL(apiPath), rdr)
	if err != nil {
		return nil, 0, err
	}
	if h != nil {
		for k, vs := range h {
			for _, v := range vs {
				req.Header.Add(k, v)
			}
		}
	}
	req.Header.Set("User-Agent", s.userAgentHeader())
	res, err := s.httpClient().Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()
	b, rerr := io.ReadAll(res.Body)
	if rerr != nil {
		return nil, res.StatusCode, rerr
	}
	return b, res.StatusCode, nil
}

func (s *State) agentBearer() string {
	return strings.TrimSpace(s.AgentAPIKey)
}

func (s *State) requireAgentKey() error {
	if s.agentBearer() == "" {
		return fmt.Errorf("AGENTGLOBE_MCP_API_KEY is not set")
	}
	return nil
}

func (s *State) agentJSON(ctx context.Context, method, apiPath string, jsonBody []byte) ([]byte, int, error) {
	if err := s.requireAgentKey(); err != nil {
		return nil, 0, err
	}
	h := make(http.Header)
	h.Set("Authorization", "Bearer "+s.agentBearer())
	if len(jsonBody) > 0 {
		h.Set("Content-Type", "application/json")
	}
	return s.httpDo(ctx, method, apiPath, h, jsonBody)
}

func (s *State) serviceRegistryJSON(ctx context.Context, method, apiPath string, jsonBody []byte) ([]byte, int, error) {
	tok := strings.TrimSpace(s.ServiceRegistryToken)
	if tok == "" {
		return nil, 0, fmt.Errorf("AGENTGLOBE_SERVICE_REGISTRY_TOKEN is not set (required for register_capability)")
	}
	h := make(http.Header)
	h.Set("Authorization", "Bearer "+tok)
	if len(jsonBody) > 0 {
		h.Set("Content-Type", "application/json")
	}
	return s.httpDo(ctx, method, apiPath, h, jsonBody)
}

func (s *State) publicGET(ctx context.Context, apiPath string) ([]byte, int, error) {
	return s.httpDo(ctx, http.MethodGet, apiPath, make(http.Header), nil)
}
