package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// capabilityServiceRegisterRequest is the JSON body for POST .../register.
type capabilityServiceRegisterRequest struct {
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
	Status       string   `json:"status"`
}

// capabilityServiceHeartbeatRequest identifies a service and optional status.
type capabilityServiceHeartbeatRequest struct {
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
	Status  string `json:"status"`
}

func capabilityServiceToMap(c *db.CapabilityService) map[string]any {
	if c == nil {
		return nil
	}
	var openAPI any
	if strings.TrimSpace(c.OpenapiSpecJSON) != "" {
		_ = json.Unmarshal([]byte(c.OpenapiSpecJSON), &openAPI)
	}
	grace := db.DefaultHeartbeatGrace
	m := map[string]any{
		"id":            c.ID,
		"name":          c.Name,
		"version":       c.Version,
		"base_url":      c.BaseURL,
		"description":   c.Description,
		"category":      c.CapabilityCategoryLabel(),
		"tags":          c.TagSlice(),
		"domains":       c.DomainsFromJSON(),
		"metadata":      c.MetadataMap(),
		"openapi_url":   c.OpenapiURL,
		"openapi_spec":  openAPI,
		"status":        c.Status,
		"is_healthy":    c.IsHealthy(grace),
		"last_seen":     c.LastSeen,
		"created_at":    c.CreatedAt,
		"updated_at":    c.UpdatedAt,
	}
	if c.CategoryID != nil {
		m["category_id"] = *c.CategoryID
	} else {
		m["category_id"] = nil
	}
	return m
}

func (s *Server) handleCapabilityServicesList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeDetail(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	qb := s.dbCtx(r).Model(&db.CapabilityService{}).Order("name ASC, base_url ASC")
	if cat := strings.TrimSpace(r.URL.Query().Get("category")); cat != "" {
		qb = qb.Where("category_id = ?", cat)
	}
	if st := strings.TrimSpace(r.URL.Query().Get("status")); st != "" {
		qb = qb.Where("LOWER(status) = LOWER(?)", st)
	}
	if search := strings.TrimSpace(r.URL.Query().Get("q")); search != "" {
		needle := strings.ToLower(search)
		like := "%" + search + "%"
		if s.dbCtx(r).Dialector.Name() == "postgres" {
			// case-insensitive substring
			qb = qb.Where(
				"name ILIKE ? OR COALESCE(description, '') ILIKE ? OR COALESCE(category_id, '') ILIKE ?",
				like, like, like,
			)
		} else {
			qb = qb.Where("instr(lower(COALESCE(name, '')), ?) > 0 OR instr(lower(COALESCE(description, '')), ?) > 0 OR instr(lower(COALESCE(category_id, '')), ?) > 0", needle, needle, needle)
		}
	}
	var rows []db.CapabilityService
	if err := qb.Preload("Category").Find(&rows).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Database error")
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, capabilityServiceToMap(&rows[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"count": len(out),
		"items": out,
	})
}

func (s *Server) handleCapabilityServiceGetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeDetail(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if id == "" {
		writeDetail(w, http.StatusBadRequest, "id is required")
		return
	}
	var row db.CapabilityService
	if err := s.dbCtx(r).Preload("Category").First(&row, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeDetail(w, http.StatusNotFound, "not found")
			return
		}
		writeDetail(w, http.StatusInternalServerError, "Database error")
		return
	}
	writeJSON(w, http.StatusOK, capabilityServiceToMap(&row))
}

func (s *Server) handleCapabilityServicesRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeDetail(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if !s.requireServiceRegistry(w, r) {
		return
	}
	var body capabilityServiceRegisterRequest
	if err := readJSON(r, &body); err != nil {
		writeDetail(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	n := strings.TrimSpace(body.Name)
	bu := strings.TrimSpace(body.BaseURL)
	ver := strings.TrimSpace(body.Version)
	if n == "" || bu == "" || ver == "" {
		writeDetail(w, http.StatusBadRequest, "name, version, and base_url are required")
		return
	}
	u, err := url.Parse(bu)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		writeDetail(w, http.StatusBadRequest, "base_url must be a valid http or https URL with a host")
		return
	}
	st := strings.TrimSpace(body.Status)
	if st == "" {
		st = db.CapabilityServiceStatusActive
	} else {
		st = strings.ToLower(st)
	}
	if !db.KnownCapabilityServiceStatus(st) {
		writeDetail(w, http.StatusBadRequest, "status must be active, degraded, or inactive")
		return
	}
	tagsJSON, _ := json.Marshal(body.Tags)
	if body.Tags == nil {
		tagsJSON = []byte("[]")
	}
	domJSON, _ := json.Marshal(body.Domains)
	if body.Domains == nil {
		domJSON = []byte("[]")
	}
	mdJSON, _ := json.Marshal(body.Metadata)
	if body.Metadata == nil {
		mdJSON = []byte("{}")
	}
	var specBytes []byte
	if body.OpenapiSpec != nil {
		var err2 error
		specBytes, err2 = json.Marshal(body.OpenapiSpec)
		if err2 != nil {
			writeDetail(w, http.StatusBadRequest, "openapi_spec is not valid JSON")
			return
		}
	}
	now := time.Now().UTC()
	catID, errCat := db.EnsureCategory(s.dbCtx(r), body.Category)
	if errCat != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not resolve category")
		return
	}
	var catPtr *string
	if catID != "" {
		catPtr = &catID
	}
	rec := db.CapabilityService{
		Name:            n,
		Version:         ver,
		BaseURL:         bu,
		Description:     body.Description,
		CategoryID:      catPtr,
		TagsJSON:        string(tagsJSON),
		DomainsJSON:     string(domJSON),
		MetadataJSON:    string(mdJSON),
		OpenapiURL:      strings.TrimSpace(body.OpenapiURL),
		OpenapiSpecJSON: string(specBytes),
		Status:          st,
		LastSeen:        &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	// id filled in BeforeCreate if still empty; upsert on (name, base_url).
	if err := s.dbCtx(r).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "name"},
			{Name: "base_url"},
		},
		DoUpdates: clause.AssignmentColumns(
			[]string{
				"version",
				"description",
				"category_id",
				"tags",
				"domains",
				"metadata",
				"openapi_url",
				"openapi_spec",
				"status",
				"last_seen",
				"updated_at",
			},
		),
	}).Create(&rec).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "register failed")
		return
	}
	// Re-load to get the stable id (insert path) and full row.
	if err := s.dbCtx(r).Preload("Category").Where("name = ? AND base_url = ?", n, bu).First(&rec).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "load after register failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"id":       rec.ID,
		"name":     rec.Name,
		"base_url": rec.BaseURL,
		"status":   rec.Status,
		"data":     capabilityServiceToMap(&rec),
	})
}

func (s *Server) handleCapabilityServicesHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeDetail(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if !s.requireServiceRegistry(w, r) {
		return
	}
	var body capabilityServiceHeartbeatRequest
	if err := readJSON(r, &body); err != nil {
		writeDetail(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	n := strings.TrimSpace(body.Name)
	bu := strings.TrimSpace(body.BaseURL)
	if n == "" || bu == "" {
		writeDetail(w, http.StatusBadRequest, "name and base_url are required")
		return
	}
	st := strings.TrimSpace(body.Status)
	if st != "" {
		st = strings.ToLower(st)
		if !db.KnownCapabilityServiceStatus(st) {
			writeDetail(w, http.StatusBadRequest, "status must be active, degraded, or inactive")
			return
		}
	}
	now := time.Now().UTC()
	updates := map[string]any{
		"last_seen":  &now,
		"updated_at": now,
	}
	if st != "" {
		updates["status"] = st
	}
	res := s.dbCtx(r).Model(&db.CapabilityService{}).Where("name = ? AND base_url = ?", n, bu).Updates(updates)
	if res.Error != nil {
		writeDetail(w, http.StatusInternalServerError, "Database error")
		return
	}
	if res.RowsAffected == 0 {
		writeDetail(w, http.StatusNotFound, "No matching capability service; register first")
		return
	}
	out := map[string]any{"ok": true, "last_seen": now}
	if st != "" {
		out["status"] = st
	}
	writeJSON(w, http.StatusOK, out)
}
