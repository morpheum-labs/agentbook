package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/credentials"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/db"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/httperr"
	"gorm.io/gorm"
)

type createCredentialBody struct {
	ProviderSlug   string          `json:"provider_slug"`
	Label          string          `json:"label"`
	McpServerName  *string         `json:"mcp_server_name"`
	Metadata       json.RawMessage `json:"metadata"`
	MaterialKind   string          `json:"material_kind"`
	Plaintext      json.RawMessage `json:"plaintext"`
}

type rotateCredentialBody struct {
	Plaintext json.RawMessage `json:"plaintext"`
}

func (s *Server) requireCredKey(w http.ResponseWriter, r *http.Request) bool {
	if len(s.credMasterKey) != 32 {
		httperr.Write(w, r, httperr.ServiceUnavailable(
			"credential writes require CLAWGOTCHA_CREDENTIALS_ENCRYPTION_KEY (32-byte raw, base64, or 64-char hex)",
		))
		return false
	}
	return true
}

func normalizeMetadata(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return json.RawMessage(`{}`)
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return json.RawMessage(`{}`)
	}
	if _, ok := v.(map[string]any); !ok {
		return json.RawMessage(`{}`)
	}
	return raw
}

func validatePlaintextJSON(raw json.RawMessage) error {
	if len(strings.TrimSpace(string(raw))) == 0 || string(raw) == "null" {
		return errors.New("plaintext is required")
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return err
	}
	return nil
}

func plaintextWireToJSONBytes(raw json.RawMessage) ([]byte, error) {
	// Store exactly what the client sent as JSON value (string or object) as UTF-8 JSON bytes.
	if err := validatePlaintextJSON(raw); err != nil {
		return nil, err
	}
	// Re-marshal to canonical compact form for encryption input.
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	return json.Marshal(v)
}

func credentialBindingToJSON(b db.CredentialBinding, latest *db.CredentialSecretVersion) map[string]any {
	out := map[string]any{
		"id":             b.ID.String(),
		"provider_slug":  b.ProviderSlug,
		"label":          b.Label,
		"metadata":       json.RawMessage(b.Metadata),
		"created_at":     b.CreatedAt.UTC().Format(time.RFC3339Nano),
		"updated_at":     b.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if b.McpServerName != nil && strings.TrimSpace(*b.McpServerName) != "" {
		out["mcp_server_name"] = strings.TrimSpace(*b.McpServerName)
	} else {
		out["mcp_server_name"] = nil
	}
	if latest != nil {
		out["current_version"] = latest.Version
		out["material_kind"] = latest.MaterialKind
		out["has_secret"] = true
		if latest.ExpiresAt != nil {
			out["expires_at"] = latest.ExpiresAt.UTC().Format(time.RFC3339Nano)
		} else {
			out["expires_at"] = nil
		}
		out["secret_updated_at"] = latest.CreatedAt.UTC().Format(time.RFC3339Nano)
	} else {
		out["current_version"] = 0
		out["material_kind"] = nil
		out["has_secret"] = false
		out["expires_at"] = nil
		out["secret_updated_at"] = nil
	}
	return out
}

func (s *Server) latestSecretVersion(bindingID uuid.UUID) (*db.CredentialSecretVersion, error) {
	var v db.CredentialSecretVersion
	err := s.db.Where("binding_id = ?", bindingID).Order("version DESC").First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (s *Server) listAgentCredentials(w http.ResponseWriter, r *http.Request) {
	agentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid id", err))
		return
	}
	if err := s.db.First(&db.SwarmAgent{}, "id = ?", agentID).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}

	var bindings []db.CredentialBinding
	if err := s.db.Where("swarm_agent_id = ?", agentID).Order("provider_slug ASC, label ASC").Find(&bindings).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	items := make([]map[string]any, 0, len(bindings))
	for i := range bindings {
		latest, err := s.latestSecretVersion(bindings[i].ID)
		if err != nil {
			httperr.Write(w, r, err)
			return
		}
		items = append(items, credentialBindingToJSON(bindings[i], latest))
	}
	sum, err := db.LoadRevisionSummary(s.db)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"credentials":      items,
		"revision_summary": sum,
	})
}

func (s *Server) createAgentCredential(w http.ResponseWriter, r *http.Request) {
	if !s.requireCredKey(w, r) {
		return
	}
	agentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid id", err))
		return
	}
	if err := s.db.First(&db.SwarmAgent{}, "id = ?", agentID).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	var b createCredentialBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid json body", err))
		return
	}
	ps := strings.TrimSpace(b.ProviderSlug)
	lb := strings.TrimSpace(b.Label)
	mk := strings.TrimSpace(b.MaterialKind)
	if ps == "" || lb == "" || mk == "" {
		httperr.Write(w, r, httperr.BadRequest("validation", errors.New("provider_slug, label, and material_kind are required")))
		return
	}
	if !credentials.ValidMaterialKind(mk) {
		httperr.Write(w, r, httperr.BadRequest("validation", errors.New("unknown material_kind")))
		return
	}
	ptBytes, err := plaintextWireToJSONBytes(b.Plaintext)
	if err != nil {
		httperr.Write(w, r, httperr.BadRequest("plaintext", err))
		return
	}
	meta := normalizeMetadata(b.Metadata)
	var mcp *string
	if b.McpServerName != nil {
		t := strings.TrimSpace(*b.McpServerName)
		if t != "" {
			mcp = &t
		}
	}

	sealed, err := credentials.Encrypt(ptBytes, s.credMasterKey)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}

	now := time.Now().UTC()
	var out db.CredentialBinding
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		bind := db.CredentialBinding{
			SwarmAgentID:  agentID,
			ProviderSlug:  ps,
			Label:         lb,
			McpServerName: mcp,
			Metadata:      meta,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := tx.Create(&bind).Error; err != nil {
			return err
		}
		ver := db.CredentialSecretVersion{
			BindingID:    bind.ID,
			Version:      1,
			MaterialKind: mk,
			Ciphertext:   sealed.Ciphertext,
			Nonce:        sealed.Nonce,
			KekID:        "env:v1",
			ExpiresAt:    nil,
			CreatedAt:    now,
		}
		if err := tx.Create(&ver).Error; err != nil {
			return err
		}
		return tx.First(&out, "id = ?", bind.ID).Error
	}); err != nil {
		if isUniqueViolation(err) {
			httperr.Write(w, r, httperr.BadRequest("duplicate binding", errors.New("credential with this provider_slug and label already exists for the agent")))
			return
		}
		httperr.Write(w, r, err)
		return
	}
	latest, err := s.latestSecretVersion(out.ID)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}
	sum, err := db.LoadRevisionSummary(s.db)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"credential":       credentialBindingToJSON(out, latest),
		"revision_summary": sum,
	})
}

func (s *Server) deleteAgentCredential(w http.ResponseWriter, r *http.Request) {
	agentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid id", err))
		return
	}
	bindingID, err := uuid.Parse(chi.URLParam(r, "bindingId"))
	if err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid bindingId", err))
		return
	}
	var bind db.CredentialBinding
	if err := s.db.First(&bind, "id = ? AND swarm_agent_id = ?", bindingID, agentID).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	if err := s.db.Delete(&bind).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	sum, err := db.LoadRevisionSummary(s.db)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":               true,
		"revision_summary": sum,
	})
}

func (s *Server) rotateAgentCredential(w http.ResponseWriter, r *http.Request) {
	if !s.requireCredKey(w, r) {
		return
	}
	agentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid id", err))
		return
	}
	bindingID, err := uuid.Parse(chi.URLParam(r, "bindingId"))
	if err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid bindingId", err))
		return
	}
	var bind db.CredentialBinding
	if err := s.db.First(&bind, "id = ? AND swarm_agent_id = ?", bindingID, agentID).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}

	var b rotateCredentialBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid json body", err))
		return
	}
	ptBytes, err := plaintextWireToJSONBytes(b.Plaintext)
	if err != nil {
		httperr.Write(w, r, httperr.BadRequest("plaintext", err))
		return
	}

	var prev db.CredentialSecretVersion
	if err := s.db.Where("binding_id = ?", bindingID).Order("version DESC").First(&prev).Error; err != nil {
		httperr.Write(w, r, httperr.BadRequest("validation", errors.New("no existing secret to rotate")))
		return
	}
	next := prev.Version + 1
	latestKind := prev.MaterialKind

	sealed, err := credentials.Encrypt(ptBytes, s.credMasterKey)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}
	now := time.Now().UTC()
	ver := db.CredentialSecretVersion{
		BindingID:    bindingID,
		Version:      next,
		MaterialKind: strings.TrimSpace(latestKind),
		Ciphertext:   sealed.Ciphertext,
		Nonce:        sealed.Nonce,
		KekID:        "env:v1",
		ExpiresAt:    nil,
		CreatedAt:    now,
	}
	if err := s.db.Create(&ver).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	if err := s.db.Model(&bind).Update("updated_at", now).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	latest, err := s.latestSecretVersion(bindingID)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}
	sum, err := db.LoadRevisionSummary(s.db)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"credential":       credentialBindingToJSON(bind, latest),
		"revision_summary": sum,
	})
}
