package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/credentials"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/db"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/httperr"
)

func (s *Server) getAgentMcpCredentialsByName(w http.ResponseWriter, r *http.Request) {
	agentName := strings.TrimSpace(chi.URLParam(r, "agent_name"))
	if agentName == "" {
		httperr.Write(w, r, httperr.BadRequest("agent_name required", nil))
		return
	}
	var agent db.SwarmAgent
	if err := s.db.Where("name = ? AND deleted_at IS NULL", agentName).First(&agent).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	s.writeAgentMcpCredentials(w, r, &agent)
}

func (s *Server) getAgentMcpCredentialsByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(chi.URLParam(r, "agent_id"))
	agentID, err := uuid.Parse(idStr)
	if err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid agent_id", err))
		return
	}
	var agent db.SwarmAgent
	if err := s.db.Where("id = ? AND deleted_at IS NULL", agentID).First(&agent).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	s.writeAgentMcpCredentials(w, r, &agent)
}

func (s *Server) writeAgentMcpCredentials(w http.ResponseWriter, r *http.Request, agent *db.SwarmAgent) {
	if len(s.credMasterKey) != 32 {
		httperr.Write(w, r, httperr.ServiceUnavailable(
			"mcp credential reveal requires CLAWGOTCHA_CREDENTIALS_ENCRYPTION_KEY (32-byte raw, base64, or 64-char hex)",
		))
		return
	}
	inst := RuntimeInstanceFromContext(r.Context())
	if inst == nil {
		httperr.Write(w, r, httperr.Forbidden("missing runtime instance context"))
		return
	}

	var bindings []db.CredentialBinding
	if err := s.db.Where("swarm_agent_id = ? AND deleted_at IS NULL", agent.ID).Order("provider_slug ASC, label ASC").Find(&bindings).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}

	mcpBindings := make([]map[string]any, 0)
	for i := range bindings {
		b := bindings[i]
		if b.McpServerName == nil || strings.TrimSpace(*b.McpServerName) == "" {
			continue
		}
		latest, err := s.latestSecretVersion(b.ID)
		if err != nil {
			httperr.Write(w, r, err)
			return
		}
		if latest == nil {
			continue
		}
		pt, err := credentials.Decrypt(&credentials.Sealed{
			Ciphertext: latest.Ciphertext,
			Nonce:      latest.Nonce,
		}, s.credMasterKey)
		if err != nil {
			httperr.Write(w, r, fmt.Errorf("decrypt: %w", err))
			return
		}
		var payload any
		if err := json.Unmarshal(pt, &payload); err != nil {
			httperr.Write(w, r, httperr.BadRequest("invalid plaintext json", err))
			return
		}
		mcpBindings = append(mcpBindings, map[string]any{
			"mcp_server_name": strings.TrimSpace(*b.McpServerName),
			"material_kind":   latest.MaterialKind,
			"payload":         payload,
		})
	}

	slog.Info("mcp_credentials_reveal",
		"instance_name", inst.InstanceName,
		"instance_id", inst.ID.String(),
		"agent_name", agent.Name,
		"agent_id", agent.ID.String(),
		"binding_count", len(mcpBindings),
	)

	writeJSON(w, http.StatusOK, map[string]any{
		"agent_name":   agent.Name,
		"mcp_bindings": mcpBindings,
		"revision":     agent.CurrentRevision,
	})
}
