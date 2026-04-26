package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"gorm.io/gorm"
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
}

// capabilityServiceHeartbeatRequest only needs identity for touch.
type capabilityServiceHeartbeatRequest struct {
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
}

func (s *Server) handleCapabilityServicesList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeDetail(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var rows []db.CapabilityService
	if err := s.dbCtx(r).Order("name ASC, base_url ASC").Find(&rows).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Database error")
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		var openAPI any
		if strings.TrimSpace(rows[i].OpenapiSpecJSON) != "" {
			_ = json.Unmarshal([]byte(rows[i].OpenapiSpecJSON), &openAPI)
		}
		out = append(out, map[string]any{
			"id":              rows[i].ID,
			"name":            rows[i].Name,
			"version":         rows[i].Version,
			"base_url":        rows[i].BaseURL,
			"description":     rows[i].Description,
			"category":        rows[i].Category,
			"tags":            rows[i].TagSlice(),
			"domains":         rows[i].DomainsFromJSON(),
			"metadata":        rows[i].MetadataMap(),
			"openapi_url":     rows[i].OpenapiURL,
			"openapi_spec":    openAPI,
			"last_seen":       rows[i].LastSeen,
			"created_at":      rows[i].CreatedAt,
			"updated_at":      rows[i].UpdatedAt,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"count": len(out),
		"items": out,
	})
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
	if _, err := url.Parse(bu); err != nil {
		writeDetail(w, http.StatusBadRequest, "base_url is not a valid URL")
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
		var err error
		specBytes, err = json.Marshal(body.OpenapiSpec)
		if err != nil {
			writeDetail(w, http.StatusBadRequest, "openapi_spec is not valid JSON")
			return
		}
	}
	now := time.Now().UTC()
	var rec db.CapabilityService
	err := s.dbCtx(r).Where("name = ? AND base_url = ?", n, bu).First(&rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		rec = db.CapabilityService{
			ID:              uuid.NewString(),
			Name:            n,
			Version:         ver,
			BaseURL:         bu,
			Description:     body.Description,
			Category:        body.Category,
			TagsJSON:        string(tagsJSON),
			DomainsJSON:     string(domJSON),
			MetadataJSON:    string(mdJSON),
			OpenapiURL:      strings.TrimSpace(body.OpenapiURL),
			OpenapiSpecJSON: string(specBytes),
			LastSeen:        &now,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := s.dbCtx(r).Create(&rec).Error; err != nil {
			writeDetail(w, http.StatusInternalServerError, "Create failed")
			return
		}
	} else if err != nil {
		writeDetail(w, http.StatusInternalServerError, "Database error")
		return
	} else {
		rec.Version = ver
		rec.Description = body.Description
		rec.Category = body.Category
		rec.TagsJSON = string(tagsJSON)
		rec.DomainsJSON = string(domJSON)
		rec.MetadataJSON = string(mdJSON)
		rec.OpenapiURL = strings.TrimSpace(body.OpenapiURL)
		rec.OpenapiSpecJSON = string(specBytes)
		rec.LastSeen = &now
		rec.UpdatedAt = now
		if err := s.dbCtx(r).Save(&rec).Error; err != nil {
			writeDetail(w, http.StatusInternalServerError, "Update failed")
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"id":     rec.ID,
		"name":   rec.Name,
		"base_url": rec.BaseURL,
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
	now := time.Now().UTC()
	res := s.dbCtx(r).Model(&db.CapabilityService{}).Where("name = ? AND base_url = ?", n, bu).Updates(map[string]any{
		"last_seen":  &now,
		"updated_at": now,
	})
	if res.Error != nil {
		writeDetail(w, http.StatusInternalServerError, "Database error")
		return
	}
	if res.RowsAffected == 0 {
		writeDetail(w, http.StatusNotFound, "No matching capability service; register first")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "last_seen": now})
}
