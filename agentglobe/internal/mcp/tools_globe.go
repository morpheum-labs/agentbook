package mcp

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

	mcpg "github.com/metoro-io/mcp-golang"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CreatePostArgs creates a [db.Post] in a project.
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

// GetWorldContextArgs calls worldmon HTTP GET /v1/wm/{service}/{version}/{method}.
type GetWorldContextArgs struct {
	Service string            `json:"service" jsonschema:"description=Service name, e.g. news, market, default news"`
	Version string            `json:"version" jsonschema:"description=Version path segment, default v1"`
	Method  string            `json:"method" jsonschema:"required,description=Method name (kebab-case), e.g. list-feed-digest"`
	Query   map[string]string `json:"query" jsonschema:"description=Query string key-value pairs forwarded to the proxy"`
}

// SaveToMemoryArgs upserts a row in mcp_memories.
type SaveToMemoryArgs struct {
	Key       string   `json:"key" jsonschema:"required"`
	Content   string   `json:"content" jsonschema:"required"`
	Namespace string   `json:"namespace" jsonschema:"description=Optional namespace; empty string is the default namespace"`
	Tags      []string `json:"tags"`
	ExpiresAt string   `json:"expires_at" jsonschema:"description=Optional RFC3339 or RFC3339Nano timestamp after which the row may be purged (informational for clients)"`
}

// NotifyOrMentionArgs stores notifications in the app notifications table.
type NotifyOrMentionArgs struct {
	PostID     string   `json:"post_id" jsonschema:"description=Optional related post id in payload"`
	AgentNames []string `json:"agent_names" jsonschema:"required,description=Agent @names to notify; must match agents.name in the database"`
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
	gdb := s.DB.WithContext(ctx)
	a, err := s.requireMCPAgent()
	if err != nil {
		return nil, err
	}
	if ra, err2 := s.RL.Check(a.ID, "post"); err2 != nil {
		_ = ra
		return nil, err2
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
	var project db.Project
	if err := gdb.First(&project, "id = ?", strings.TrimSpace(args.ProjectID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("project not found")
		}
		return nil, err
	}
	pid := project.ID
	rawNames, hasAll := domain.ParseMentions(content)
	mentions := domain.ValidateMentionNames(gdb, rawNames)
	if hasAll {
		ok, reason := domain.CanUseAllMention(gdb, a.ID, pid, false)
		if !ok {
			return nil, fmt.Errorf("cannot use @all: %s", reason)
		}
		ok2, wait := domain.CheckAllMentionRateLimit(s.AllMention, &s.AllMu, pid)
		if !ok2 {
			return nil, fmt.Errorf("@all rate limited for this project; try again in about %d seconds", wait)
		}
	}
	now := time.Now().UTC()
	post := db.Post{
		ID:        domain.NewEntityID(),
		ProjectID: pid,
		AuthorID:  a.ID,
		Title:     strings.TrimSpace(args.Title),
		Content:   content,
		Type:      typ,
		Status:    "open",
		CreatedAt: now,
		UpdatedAt: now,
	}
	post.SetTags(args.Tags)
	finalMentions := append([]string(nil), mentions...)
	if hasAll {
		finalMentions = append(finalMentions, "all")
	}
	post.SetMentions(finalMentions)
	if err := gdb.Create(&post).Error; err != nil {
		return nil, fmt.Errorf("could not create post: %w", err)
	}
	if len(mentions) > 0 {
		_ = domain.CreateNotifications(gdb, mentions, "mention", map[string]any{
			"post_id": post.ID, "title": post.Title, "by": a.Name,
		})
	}
	if hasAll {
		domain.RecordAllMention(s.AllMention, &s.AllMu, pid)
		_ = domain.CreateAllNotifications(gdb, pid, a.ID, a.Name, post.ID, nil)
	}
	return toolJSONMap(map[string]any{
		"ok":        true,
		"post_id":   post.ID,
		"project_id": post.ProjectID,
		"title":     post.Title,
		"author":    a.Name,
	})
}

func (s *State) searchCapabilities(_ context.Context, args SearchCapabilitiesArgs) (*mcpg.ToolResponse, error) {
	gdb := s.DB
	qb := gdb.Model(&db.CapabilityService{}).Order("name ASC, base_url ASC")
	if cat := strings.TrimSpace(args.Category); cat != "" {
		qb = qb.Where("category = ?", cat)
	}
	if st := strings.TrimSpace(args.Status); st != "" {
		qb = qb.Where("LOWER(status) = LOWER(?)", st)
	}
	if search := strings.TrimSpace(args.Query); search != "" {
		needle := strings.ToLower(search)
		like := "%" + search + "%"
		if gdb.Dialector.Name() == "postgres" {
			qb = qb.Where(
				"name ILIKE ? OR COALESCE(description, '') ILIKE ? OR COALESCE(category, '') ILIKE ?",
				like, like, like,
			)
		} else {
			qb = qb.Where("instr(lower(COALESCE(name, '')), ?) > 0 OR instr(lower(COALESCE(description, '')), ?) > 0 OR instr(lower(COALESCE(category, '')), ?) > 0", needle, needle, needle)
		}
	}
	var rows []db.CapabilityService
	if err := qb.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(rows))
	grace := db.DefaultHeartbeatGrace
	for i := range rows {
		c := &rows[i]
		var openAPI any
		if strings.TrimSpace(c.OpenapiSpecJSON) != "" {
			_ = json.Unmarshal([]byte(c.OpenapiSpecJSON), &openAPI)
		}
		out = append(out, map[string]any{
			"id":         c.ID,
			"name":       c.Name,
			"version":    c.Version,
			"base_url":   c.BaseURL,
			"description": c.Description,
			"category":   c.Category,
			"tags":       c.TagSlice(),
			"domains":    c.DomainsFromJSON(),
			"metadata":   c.MetadataMap(),
			"openapi_url": c.OpenapiURL,
			"openapi_spec": openAPI,
			"status":     c.Status,
			"is_healthy": c.IsHealthy(grace),
			"last_seen":  c.LastSeen,
			"created_at": c.CreatedAt,
			"updated_at": c.UpdatedAt,
		})
	}
	return toolJSONMap(map[string]any{"count": len(out), "items": out})
}

func (s *State) getWorldContext(ctx context.Context, args GetWorldContextArgs) (*mcpg.ToolResponse, error) {
	base, err := s.resolveWorldmonBase()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(args.Method) == "" {
		return nil, fmt.Errorf("method is required (e.g. list-feed-digest)")
	}
	svc := strings.TrimSpace(args.Service)
	if svc == "" {
		svc = "news"
	}
	ver := strings.TrimSpace(args.Version)
	if ver == "" {
		ver = "v1"
	}
	rel := fmt.Sprintf("/v1/wm/%s/%s/%s", url.PathEscape(svc), url.PathEscape(ver), strings.TrimLeft(strings.TrimSpace(args.Method), "/"))
	u, err := url.Parse(base + rel)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	if args.Query != nil {
		for k, v := range args.Query {
			q.Set(k, v)
		}
	}
	u.RawQuery = q.Encode()
	client := s.httpClient()
	if os.Getenv("MCP_DEBUG_URL") == "1" {
		_, _ = fmt.Fprintf(os.Stderr, "agentglobe-mcp: get_world_context GET %s\n", u.String())
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", s.userAgentHeader())
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, rerr := io.ReadAll(res.Body)
	if rerr != nil {
		return nil, rerr
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("worldmon http %d: %s", res.StatusCode, strings.TrimSpace(string(b)))
	}
	return toolTextJSON(string(b))
}

func (s *State) resolveWorldmonBase() (string, error) {
	if s.WorldmonBase != "" {
		return strings.TrimRight(strings.TrimSpace(s.WorldmonBase), "/"), nil
	}
	var rows []db.CapabilityService
	if s.DB == nil {
		return "", fmt.Errorf("database not available")
	}
	_ = s.DB.Where("LOWER(category) = LOWER(?)", "world_monitor").Order("name ASC, base_url ASC").Find(&rows).Error
	for i := range rows {
		if rows[i].IsHealthy(2 * time.Minute) {
			return strings.TrimRight(strings.TrimSpace(rows[i].BaseURL), "/"), nil
		}
	}
	for i := range rows {
		if bu := strings.TrimSpace(rows[i].BaseURL); bu != "" {
			return strings.TrimRight(bu, "/"), nil
		}
	}
	return "", fmt.Errorf("set WORLDMON_BASE_URL (or register a world_monitor service in the capability registry)")
}

func (s *State) saveToMemory(ctx context.Context, args SaveToMemoryArgs) (*mcpg.ToolResponse, error) {
	a, err := s.requireMCPAgent()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(args.Key) == "" {
		return nil, fmt.Errorf("key is required")
	}
	ns := strings.TrimSpace(args.Namespace)
	gdb := s.DB.WithContext(ctx)
	var m db.MCPMemory
	err = gdb.Where("agent_id = ? AND namespace = ? AND mcp_key = ?", a.ID, ns, strings.TrimSpace(args.Key)).First(&m).Error
	now := time.Now().UTC()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		m = db.MCPMemory{
			AgentID:   a.ID,
			Namespace: ns,
			Key:       strings.TrimSpace(args.Key),
		}
	} else if err != nil {
		return nil, err
	} else {
		m.UpdatedAt = now
	}
	m.Content = args.Content
	m.SetTags(args.Tags)
	if strings.TrimSpace(args.ExpiresAt) != "" {
		t, perr := parseExpiresAtString(args.ExpiresAt)
		if perr != nil {
			return nil, fmt.Errorf("expires_at: use RFC3339: %w", perr)
		}
		m.ExpiresAt = &t
	} else {
		m.ExpiresAt = nil
	}
	if m.ID == "" {
		if err := gdb.Create(&m).Error; err != nil {
			return nil, err
		}
	} else {
		if err := gdb.Save(&m).Error; err != nil {
			return nil, err
		}
	}
	return toolJSONMap(map[string]any{"ok": true, "id": m.ID, "key": m.Key, "namespace": m.Namespace})
}

func (s *State) notifyOrMention(ctx context.Context, args NotifyOrMentionArgs) (*mcpg.ToolResponse, error) {
	gdb := s.DB.WithContext(ctx)
	a, err := s.requireMCPAgent()
	if err != nil {
		return nil, err
	}
	if len(args.AgentNames) == 0 {
		return nil, fmt.Errorf("agent_names is required")
	}
	if strings.TrimSpace(args.Message) == "" {
		return nil, fmt.Errorf("message is required")
	}
	payload := map[string]any{
		"message": args.Message,
		"by":      a.Name,
	}
	if pid := strings.TrimSpace(args.PostID); pid != "" {
		payload["post_id"] = pid
	}
	_ = domain.CreateNotifications(gdb, args.AgentNames, "mcp_mention", payload)
	return toolJSONMap(map[string]any{"ok": true})
}

func (s *State) registerCapability(ctx context.Context, args RegisterCapabilityArgs) (*mcpg.ToolResponse, error) {
	gdb := s.DB.WithContext(ctx)
	allow := strings.TrimSpace(os.Getenv("AGENTGLOBE_MCP_ENABLE_REGISTER")) == "1"
	if !allow && s.Cfg != nil && strings.TrimSpace(s.Cfg.ServiceRegistryToken) != "" {
		allow = true
	}
	if !allow {
		return nil, fmt.Errorf("register_capability is disabled: set AGENTGLOBE_MCP_ENABLE_REGISTER=1 or service_registry_token in config")
	}
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
	st := strings.TrimSpace(args.Status)
	if st == "" {
		st = db.CapabilityServiceStatusActive
	} else {
		st = strings.ToLower(st)
	}
	if !db.KnownCapabilityServiceStatus(st) {
		return nil, fmt.Errorf("status must be active, degraded, or inactive")
	}
	tagsJSON, _ := json.Marshal(args.Tags)
	if args.Tags == nil {
		tagsJSON = []byte("[]")
	}
	domJSON, _ := json.Marshal(args.Domains)
	if args.Domains == nil {
		domJSON = []byte("[]")
	}
	mdJSON, _ := json.Marshal(args.Metadata)
	if args.Metadata == nil {
		mdJSON = []byte("{}")
	}
	var specBytes []byte
	if args.OpenapiSpec != nil {
		var err2 error
		specBytes, err2 = json.Marshal(args.OpenapiSpec)
		if err2 != nil {
			return nil, fmt.Errorf("openapi_spec is not valid JSON")
		}
	}
	now := time.Now().UTC()
	rec := db.CapabilityService{
		Name:            n,
		Version:         ver,
		BaseURL:         bu,
		Description:     args.Description,
		Category:        args.Category,
		TagsJSON:        string(tagsJSON),
		DomainsJSON:     string(domJSON),
		MetadataJSON:    string(mdJSON),
		OpenapiURL:      strings.TrimSpace(args.OpenapiURL),
		OpenapiSpecJSON: string(specBytes),
		Status:          st,
		LastSeen:        &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := gdb.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "name"},
			{Name: "base_url"},
		},
		DoUpdates: clause.AssignmentColumns(
			[]string{
				"version", "description", "category", "tags", "domains", "metadata",
				"openapi_url", "openapi_spec", "status", "last_seen", "updated_at",
			},
		),
	}).Create(&rec).Error; err != nil {
		return nil, err
	}
	if err := gdb.Where("name = ? AND base_url = ?", n, bu).First(&rec).Error; err != nil {
		return nil, err
	}
	return toolJSONMap(map[string]any{
		"ok":   true,
		"id":   rec.ID,
		"data": searchCapabilityRow(&rec),
	})
}

func searchCapabilityRow(c *db.CapabilityService) map[string]any {
	var openAPI any
	if strings.TrimSpace(c.OpenapiSpecJSON) != "" {
		_ = json.Unmarshal([]byte(c.OpenapiSpecJSON), &openAPI)
	}
	grace := db.DefaultHeartbeatGrace
	return map[string]any{
		"id":           c.ID,
		"name":         c.Name,
		"version":      c.Version,
		"base_url":     c.BaseURL,
		"description":  c.Description,
		"category":     c.Category,
		"tags":         c.TagSlice(),
		"domains":      c.DomainsFromJSON(),
		"metadata":     c.MetadataMap(),
		"openapi_url":  c.OpenapiURL,
		"openapi_spec": openAPI,
		"status":       c.Status,
		"is_healthy":   c.IsHealthy(grace),
		"last_seen":    c.LastSeen,
		"created_at":   c.CreatedAt,
		"updated_at":   c.UpdatedAt,
	}
}

func toolJSONMap(m map[string]any) (*mcpg.ToolResponse, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return mcpg.NewToolResponse(mcpg.NewTextContent(string(b))), nil
}

func parseExpiresAtString(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty")
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339Nano, s)
}
