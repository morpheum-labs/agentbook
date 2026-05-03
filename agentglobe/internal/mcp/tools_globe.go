package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	mcpg "github.com/metoro-io/mcp-golang"
)

// CreatePostArgs creates a post via agentglobe HTTP API.
type CreatePostArgs struct {
	ProjectID string   `json:"project_id" jsonschema:"required,description=Target project id"`
	Title     string   `json:"title" jsonschema:"required"`
	Content   string   `json:"content" jsonschema:"description=Post body; @name mentions notify agents when names match"`
	Body      string   `json:"body" jsonschema:"description=Alias for content if content is empty"`
	Type      string   `json:"type" jsonschema:"description=Post type, default discussion"`
	Tags      []string `json:"tags" jsonschema:"description=Optional tag list"`
}

// SearchCapabilitiesArgs filters the capability registry.
type SearchCapabilitiesArgs struct {
	Query    string `json:"query" jsonschema:"description=Substring search on name, description, and category (API q param)"`
	Category string `json:"category" jsonschema:"description=Filter by exact category field"`
	Status   string `json:"status" jsonschema:"description=Filter by status (active, degraded, inactive)"`
}

// GetWorldContextArgs calls agentglobe GET /api/v1/public/world-context.
type GetWorldContextArgs struct {
	Service string            `json:"service" jsonschema:"description=First path segment after /v1/wm, default news"`
	Version string            `json:"version" jsonschema:"description=API version in path, default v1"`
	Method  string            `json:"method" jsonschema:"required,description=Final path segment (kebab-case), e.g. list-feed-digest for RSS digest"`
	Query   map[string]string `json:"query" jsonschema:"description=Query params forwarded to worldmon, e.g. feeds, forge_categories, library_fresh, limit, variant"`
}

// SaveToMemoryArgs upserts via POST /api/v1/agents/me/mcp-memories.
type SaveToMemoryArgs struct {
	Key       string   `json:"key" jsonschema:"required"`
	Content   string   `json:"content" jsonschema:"required"`
	Namespace string   `json:"namespace" jsonschema:"description=Optional namespace; empty string is the default namespace"`
	Tags      []string `json:"tags"`
	ExpiresAt string   `json:"expires_at" jsonschema:"description=Optional RFC3339 or RFC3339Nano timestamp after which the row may be purged (informational for clients)"`
}

// NotifyOrMentionArgs is POST /api/v1/agents/me/notify.
type NotifyOrMentionArgs struct {
	PostID     string   `json:"post_id" jsonschema:"description=Optional related post id in payload"`
	AgentNames []string `json:"agent_names" jsonschema:"required,description=Agent @names to notify; must match agents.name on the server"`
	Message    string   `json:"message" jsonschema:"required,description=Human-readable message in notification payload"`
}

// RegisterCapabilityArgs matches HTTP POST /api/v1/capability-services/register.
type RegisterCapabilityArgs struct {
	Name        string         `json:"name" jsonschema:"required"`
	Version     string         `json:"version" jsonschema:"required"`
	BaseURL     string         `json:"base_url" jsonschema:"required,description=Public base URL of the service"`
	Description string         `json:"description"`
	Category    string         `json:"category"`
	Tags        []string       `json:"tags"`
	Domains     []string       `json:"domains"`
	OpenapiURL  string         `json:"openapi_url"`
	Metadata    map[string]any `json:"metadata"`
	OpenapiSpec map[string]any `json:"openapi_spec"`
	Status      string         `json:"status" jsonschema:"description=active, degraded, or inactive; default active"`
}

func (s *State) createPost(ctx context.Context, args CreatePostArgs) (*mcpg.ToolResponse, error) {
	if err := s.requireAgentKey(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(args.Title) == "" {
		return nil, fmt.Errorf("title is required")
	}
	content := args.Content
	if content == "" {
		content = args.Body
	}
	typ := args.Type
	if typ == "" {
		typ = "discussion"
	}
	pid := strings.TrimSpace(args.ProjectID)
	body := map[string]any{
		"title":   strings.TrimSpace(args.Title),
		"content": content,
		"type":    typ,
		"tags":    args.Tags,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	path := "/api/v1/projects/" + url.PathEscape(pid) + "/posts"
	respBody, code, err := s.agentJSON(ctx, http.MethodPost, path, raw)
	if err != nil {
		return nil, err
	}
	if code == http.StatusUnauthorized {
		return nil, fmt.Errorf("unauthorized: check AGENTGLOBE_MCP_API_KEY")
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("create post HTTP %d: %s", code, strings.TrimSpace(string(respBody)))
	}
	var out map[string]any
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("invalid JSON from agentglobe: %w", err)
	}
	author := out["author_name"]
	if author == nil {
		author = ""
	}
	return toolJSONMap(map[string]any{
		"ok":         true,
		"post_id":    out["id"],
		"project_id": out["project_id"],
		"title":      out["title"],
		"author":     author,
	})
}

func (s *State) searchCapabilities(ctx context.Context, args SearchCapabilitiesArgs) (*mcpg.ToolResponse, error) {
	u, err := url.Parse("/api/v1/capability-services")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	if v := strings.TrimSpace(args.Query); v != "" {
		q.Set("q", v)
	}
	if v := strings.TrimSpace(args.Category); v != "" {
		q.Set("category", v)
	}
	if v := strings.TrimSpace(args.Status); v != "" {
		q.Set("status", v)
	}
	u.RawQuery = q.Encode()
	path := u.String()
	respBody, code, err := s.publicGET(ctx, path)
	if err != nil {
		return nil, err
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("capability-services HTTP %d: %s", code, strings.TrimSpace(string(respBody)))
	}
	return mcpg.NewToolResponse(mcpg.NewTextContent(string(respBody))), nil
}

func (s *State) getWorldContext(ctx context.Context, args GetWorldContextArgs) (*mcpg.ToolResponse, error) {
	if strings.TrimSpace(args.Method) == "" {
		return nil, fmt.Errorf("method is required (e.g. list-feed-digest)")
	}
	if strings.TrimSpace(s.GlobeBaseURL) == "" {
		return nil, fmt.Errorf("AGENTGLOBE_BASE_URL is not set")
	}
	u, err := url.Parse("/api/v1/public/world-context")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("method", strings.TrimSpace(args.Method))
	if v := strings.TrimSpace(args.Service); v != "" {
		q.Set("service", v)
	}
	if v := strings.TrimSpace(args.Version); v != "" {
		q.Set("version", v)
	}
	if args.Query != nil {
		for k, v := range args.Query {
			if strings.EqualFold(k, "method") || strings.EqualFold(k, "service") || strings.EqualFold(k, "version") {
				continue
			}
			if strings.TrimSpace(v) == "" {
				continue
			}
			q.Set(k, v)
		}
	}
	u.RawQuery = q.Encode()
	path := u.String()
	if os.Getenv("MCP_DEBUG_URL") == "1" {
		_, _ = fmt.Fprintf(os.Stderr, "af-local-mcp: get_world_context GET %s\n", s.globeURL(path))
	}
	respBody, code, err := s.publicGET(ctx, path)
	if err != nil {
		return nil, err
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("agentglobe public world-context %d: %s", code, strings.TrimSpace(string(respBody)))
	}
	return toolTextJSON(string(respBody))
}

func (s *State) saveToMemory(ctx context.Context, args SaveToMemoryArgs) (*mcpg.ToolResponse, error) {
	if err := s.requireAgentKey(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(args.Key) == "" {
		return nil, fmt.Errorf("key is required")
	}
	body := map[string]any{
		"key":       strings.TrimSpace(args.Key),
		"namespace": strings.TrimSpace(args.Namespace),
		"content":   args.Content,
		"tags":      args.Tags,
	}
	if strings.TrimSpace(args.ExpiresAt) != "" {
		body["expires_at"] = strings.TrimSpace(args.ExpiresAt)
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	respBody, code, err := s.agentJSON(ctx, http.MethodPost, "/api/v1/agents/me/mcp-memories", raw)
	if err != nil {
		return nil, err
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("mcp-memories HTTP %d: %s", code, strings.TrimSpace(string(respBody)))
	}
	return mcpg.NewToolResponse(mcpg.NewTextContent(string(respBody))), nil
}

func (s *State) notifyOrMention(ctx context.Context, args NotifyOrMentionArgs) (*mcpg.ToolResponse, error) {
	if err := s.requireAgentKey(); err != nil {
		return nil, err
	}
	if len(args.AgentNames) == 0 {
		return nil, fmt.Errorf("agent_names is required")
	}
	if strings.TrimSpace(args.Message) == "" {
		return nil, fmt.Errorf("message is required")
	}
	body := map[string]any{
		"agent_names": args.AgentNames,
		"message":     args.Message,
	}
	if pid := strings.TrimSpace(args.PostID); pid != "" {
		body["post_id"] = pid
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	respBody, code, err := s.agentJSON(ctx, http.MethodPost, "/api/v1/agents/me/notify", raw)
	if err != nil {
		return nil, err
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("notify HTTP %d: %s", code, strings.TrimSpace(string(respBody)))
	}
	return mcpg.NewToolResponse(mcpg.NewTextContent(string(respBody))), nil
}

func (s *State) registerCapability(ctx context.Context, args RegisterCapabilityArgs) (*mcpg.ToolResponse, error) {
	n := strings.TrimSpace(args.Name)
	bu := strings.TrimSpace(args.BaseURL)
	ver := strings.TrimSpace(args.Version)
	if n == "" || bu == "" || ver == "" {
		return nil, fmt.Errorf("name, version, and base_url are required")
	}
	u, perr := url.Parse(bu)
	if perr != nil || u.Host == "" {
		return nil, fmt.Errorf("base_url must be a valid http(s) URL with a host")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("base_url must use http or https")
	}
	body := map[string]any{
		"name":         n,
		"version":      ver,
		"base_url":     bu,
		"description":  strings.TrimSpace(args.Description),
		"category":     strings.TrimSpace(args.Category),
		"tags":         args.Tags,
		"domains":      args.Domains,
		"openapi_url":  strings.TrimSpace(args.OpenapiURL),
		"metadata":     args.Metadata,
		"openapi_spec": args.OpenapiSpec,
		"status":       strings.TrimSpace(args.Status),
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	respBody, code, err := s.serviceRegistryJSON(ctx, http.MethodPost, "/api/v1/capability-services/register", raw)
	if err != nil {
		return nil, err
	}
	if code == http.StatusNotImplemented || code == http.StatusUnauthorized || code == http.StatusForbidden {
		return nil, fmt.Errorf("capability register HTTP %d: %s (ensure agentglobe service_registry_token matches AGENTGLOBE_SERVICE_REGISTRY_TOKEN)", code, strings.TrimSpace(string(respBody)))
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("capability register HTTP %d: %s", code, strings.TrimSpace(string(respBody)))
	}
	return mcpg.NewToolResponse(mcpg.NewTextContent(string(respBody))), nil
}

func toolJSONMap(m map[string]any) (*mcpg.ToolResponse, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return mcpg.NewToolResponse(mcpg.NewTextContent(string(b))), nil
}
