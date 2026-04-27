package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/morpheumlabs/agentbook/clawlaundry/internal/db"
	"github.com/morpheumlabs/agentbook/clawlaundry/internal/httperr"
	"gorm.io/gorm"
)

type createAgentBody struct {
	Name            string  `json:"name"`
	SystemPrompt    string  `json:"system_prompt"`
	Identity        *string `json:"identity"`
	Soul            *string `json:"soul"`
	UserContext     *string `json:"user_context"`
	Tools           []string `json:"tools"`
	Provider        string  `json:"provider"`
	Model           string  `json:"model"`
	TimeoutSeconds  int     `json:"timeout_seconds"`
	AutonomyLevel   string  `json:"autonomy_level"`
}

type patchAgentBody struct {
	Name            *string `json:"name"`
	SystemPrompt    *string `json:"system_prompt"`
	Identity        *string `json:"identity"`
	Soul            *string `json:"soul"`
	UserContext     *string `json:"user_context"`
	Tools           []string `json:"tools,omitempty"`
	Provider        *string `json:"provider"`
	Model           *string `json:"model"`
	TimeoutSeconds  *int    `json:"timeout_seconds"`
	AutonomyLevel   *string `json:"autonomy_level"`
	ClearTools      *bool   `json:"clear_tools"`
}

func (s *Server) listAgents(w http.ResponseWriter, r *http.Request) {
	var out []db.SwarmAgent
	if err := s.db.Order("name").Find(&out).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	if out == nil {
		out = []db.SwarmAgent{}
	}
	agents := make([]swarmAgentResponse, 0, len(out))
	for i := range out {
		agents = append(agents, toSwarmAgentResponse(out[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{"agents": agents})
}

func (s *Server) createAgent(w http.ResponseWriter, r *http.Request) {
	var b createAgentBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid json body", err))
		return
	}
	if strings.TrimSpace(b.Name) == "" {
		httperr.Write(w, r, httperr.BadRequest("name required", nil))
		return
	}
	if err := validateAutonomy(b.AutonomyLevel); err != nil {
		httperr.Write(w, r, httperr.BadRequest("validation", err))
		return
	}
	sp := b.resolvedSystemPrompt()
	agent := db.SwarmAgent{
		Name:            strings.TrimSpace(b.Name),
		SystemPrompt:    sp,
		Tools:           b.Tools,
		Provider:        b.Provider,
		Model:           b.Model,
		TimeoutSeconds:  b.TimeoutSeconds,
		AutonomyLevel:   strings.TrimSpace(b.AutonomyLevel),
	}
	if agent.Tools == nil {
		agent.Tools = []string{}
	}
	if err := s.db.Create(&agent).Error; err != nil {
		if isUniqueViolation(err) {
			httperr.Write(w, r, httperr.BadRequest("duplicate", err))
			return
		}
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toSwarmAgentResponse(agent))
}

func (s *Server) getAgent(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid id", err))
		return
	}
	var a db.SwarmAgent
	if err := s.db.First(&a, "id = ?", id).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toSwarmAgentResponse(a))
}

func (s *Server) getAgentByName(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(chi.URLParam(r, "name"))
	if name == "" {
		httperr.Write(w, r, httperr.BadRequest("name required", nil))
		return
	}
	var a db.SwarmAgent
	if err := s.db.Where("name = ?", name).First(&a).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toSwarmAgentResponse(a))
}

func (s *Server) putAgent(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid id", err))
		return
	}
	var b createAgentBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid json body", err))
		return
	}
	if strings.TrimSpace(b.Name) == "" {
		httperr.Write(w, r, httperr.BadRequest("name required", nil))
		return
	}
	if err := validateAutonomy(b.AutonomyLevel); err != nil {
		httperr.Write(w, r, httperr.BadRequest("validation", err))
		return
	}
	sp := b.resolvedSystemPrompt()
	agent := db.SwarmAgent{
		ID:             id,
		Name:           strings.TrimSpace(b.Name),
		SystemPrompt:   sp,
		Tools:          b.Tools,
		Provider:       b.Provider,
		Model:          b.Model,
		TimeoutSeconds: b.TimeoutSeconds,
		AutonomyLevel:  strings.TrimSpace(b.AutonomyLevel),
	}
	if agent.Tools == nil {
		agent.Tools = []string{}
	}
	// Full replace: upsert on primary key, conflict on name if renamed to another row — detect
	var existing db.SwarmAgent
	if err := s.db.First(&existing, "id = ?", id).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	agent.ID = existing.ID
	agent.CreatedAt = existing.CreatedAt
	if err := s.db.Save(&agent).Error; err != nil {
		if isUniqueViolation(err) {
			httperr.Write(w, r, httperr.BadRequest("duplicate name", err))
			return
		}
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toSwarmAgentResponse(agent))
}

func (s *Server) patchAgent(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid id", err))
		return
	}
	var b patchAgentBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid json body", err))
		return
	}
	var a db.SwarmAgent
	if err := s.db.First(&a, "id = ?", id).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	updates := map[string]any{}
	if b.Name != nil {
		updates["name"] = strings.TrimSpace(*b.Name)
	}
	if v, ok, err := applyPatchSystemPrompt(a.SystemPrompt, &b); err != nil {
		httperr.Write(w, r, httperr.BadRequest("system_prompt", err))
		return
	} else if ok {
		updates["system_prompt"] = v
	}
	if b.Tools != nil {
		updates["tools"] = b.Tools
	}
	if b.ClearTools != nil && *b.ClearTools {
		updates["tools"] = []string{}
	}
	if b.Provider != nil {
		updates["provider"] = *b.Provider
	}
	if b.Model != nil {
		updates["model"] = *b.Model
	}
	if b.TimeoutSeconds != nil {
		updates["timeout_seconds"] = *b.TimeoutSeconds
	}
	if b.AutonomyLevel != nil {
		if err := validateAutonomy(*b.AutonomyLevel); err != nil {
			httperr.Write(w, r, httperr.BadRequest("validation", err))
			return
		}
		updates["autonomy_level"] = strings.TrimSpace(*b.AutonomyLevel)
	}
	if len(updates) == 0 {
		writeJSON(w, http.StatusOK, a)
		return
	}
	if err := s.db.Model(&a).Updates(updates).Error; err != nil {
		if isUniqueViolation(err) {
			httperr.Write(w, r, httperr.BadRequest("duplicate name", err))
			return
		}
		httperr.Write(w, r, err)
		return
	}
	if err := s.db.First(&a, "id = ?", id).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toSwarmAgentResponse(a))
}

func (s *Server) deleteAgent(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid id", err))
		return
	}
	tx := s.db.Delete(&db.SwarmAgent{}, "id = ?", id)
	if err := tx.Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	if tx.RowsAffected == 0 {
		httperr.Write(w, r, gorm.ErrRecordNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
