package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/db"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/events"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/httperr"
	"gorm.io/gorm"
)

type createAgentBody struct {
	Name           string   `json:"name"`
	SystemPrompt   string   `json:"system_prompt"`
	Identity       *string  `json:"identity"`
	Soul           *string  `json:"soul"`
	UserContext    *string  `json:"user_context"`
	Tools          []string `json:"tools"`
	Provider       string   `json:"provider"`
	Model          string   `json:"model"`
	TimeoutSeconds int      `json:"timeout_seconds"`
	AutonomyLevel  string   `json:"autonomy_level"`
}

type patchAgentBody struct {
	Name           *string  `json:"name"`
	SystemPrompt   *string  `json:"system_prompt"`
	Identity       *string  `json:"identity"`
	Soul           *string  `json:"soul"`
	UserContext    *string  `json:"user_context"`
	Tools          []string `json:"tools,omitempty"`
	Provider       *string  `json:"provider"`
	Model          *string  `json:"model"`
	TimeoutSeconds *int     `json:"timeout_seconds"`
	AutonomyLevel  *string  `json:"autonomy_level"`
	ClearTools     *bool    `json:"clear_tools"`
}

func (s *Server) listAgents(w http.ResponseWriter, r *http.Request) {
	sinceRev, updatedAfter, delta, qerr := parseRevisionQuery(r)
	if qerr != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid query", qerr))
		return
	}
	tx := s.db
	if delta {
		tx = tx.Unscoped()
	}
	tx = tx.Order("name")
	if sinceRev > 0 {
		tx = tx.Where("current_revision > ?", sinceRev)
	}
	if updatedAfter != nil {
		tx = tx.Where("last_changed_at > ?", *updatedAfter)
	}
	var out []db.SwarmAgent
	if err := tx.Find(&out).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	if out == nil {
		out = []db.SwarmAgent{}
	}
	sum, err := db.LoadRevisionSummary(s.db)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}
	agents := make([]swarmAgentResponse, 0, len(out))
	for i := range out {
		agents = append(agents, toSwarmAgentResponse(out[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"agents":           agents,
		"revision_summary": sum,
	})
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
	db.TouchAgentRevision(&agent)
	if err := s.db.Create(&agent).Error; err != nil {
		if isUniqueViolation(err) {
			httperr.Write(w, r, httperr.BadRequest("duplicate", err))
			return
		}
		httperr.Write(w, r, err)
		return
	}
	s.emit(events.ChangeEvent{
		EventType:          events.EventAgentUpdated,
		AffectedEntityType: events.EntityAgent,
		AffectedIDs:        []string{agent.ID.String()},
		NewRevision:        agent.CurrentRevision,
		TS:                 events.NowRFC3339Nano(),
	})
	sum, err := db.LoadRevisionSummary(s.db)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"agent":            toSwarmAgentResponse(agent),
		"revision_summary": sum,
	})
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
	sum, err := db.LoadRevisionSummary(s.db)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"agent":            toSwarmAgentResponse(a),
		"revision_summary": sum,
	})
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
	sum, err := db.LoadRevisionSummary(s.db)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"agent":            toSwarmAgentResponse(a),
		"revision_summary": sum,
	})
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
	var existing db.SwarmAgent
	if err := s.db.First(&existing, "id = ?", id).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	agent := db.SwarmAgent{
		ID:              existing.ID,
		CreatedAt:       existing.CreatedAt,
		Name:            strings.TrimSpace(b.Name),
		SystemPrompt:    sp,
		Tools:           b.Tools,
		Provider:        b.Provider,
		Model:           b.Model,
		TimeoutSeconds:  b.TimeoutSeconds,
		AutonomyLevel:   strings.TrimSpace(b.AutonomyLevel),
		CurrentRevision: existing.CurrentRevision,
	}
	if agent.Tools == nil {
		agent.Tools = []string{}
	}
	var out db.SwarmAgent
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&agent).Error; err != nil {
			return err
		}
		if err := db.IncrementAgentRevision(tx, agent.ID); err != nil {
			return err
		}
		return tx.First(&out, "id = ?", agent.ID).Error
	})
	if err != nil {
		if isUniqueViolation(err) {
			httperr.Write(w, r, httperr.BadRequest("duplicate name", err))
			return
		}
		httperr.Write(w, r, err)
		return
	}
	s.emit(events.ChangeEvent{
		EventType:          events.EventAgentUpdated,
		AffectedEntityType: events.EntityAgent,
		AffectedIDs:        []string{out.ID.String()},
		NewRevision:        out.CurrentRevision,
		TS:                 events.NowRFC3339Nano(),
	})
	sum, err := db.LoadRevisionSummary(s.db)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"agent":            toSwarmAgentResponse(out),
		"revision_summary": sum,
	})
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
		sum, err := db.LoadRevisionSummary(s.db)
		if err != nil {
			httperr.Write(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"agent":            toSwarmAgentResponse(a),
			"revision_summary": sum,
		})
		return
	}
	now := time.Now().UTC()
	updates["current_revision"] = gorm.Expr("current_revision + 1")
	updates["last_changed_at"] = now
	updates["updated_at"] = now
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
	s.emit(events.ChangeEvent{
		EventType:          events.EventAgentUpdated,
		AffectedEntityType: events.EntityAgent,
		AffectedIDs:        []string{a.ID.String()},
		NewRevision:        a.CurrentRevision,
		TS:                 events.NowRFC3339Nano(),
	})
	sum, err := db.LoadRevisionSummary(s.db)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"agent":            toSwarmAgentResponse(a),
		"revision_summary": sum,
	})
}

func (s *Server) deleteAgent(w http.ResponseWriter, r *http.Request) {
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
	now := time.Now().UTC()
	if err := s.db.Model(&db.SwarmAgent{}).Unscoped().Where("id = ?", id).Updates(map[string]any{
		"deleted_at":       now,
		"current_revision": gorm.Expr("current_revision + 1"),
		"last_changed_at":  now,
	}).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	if err := s.db.Unscoped().First(&a, "id = ?", id).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	s.emit(events.ChangeEvent{
		EventType:          events.EventAgentDeleted,
		AffectedEntityType: events.EntityAgent,
		AffectedIDs:        []string{a.ID.String()},
		NewRevision:        a.CurrentRevision,
		TS:                 events.NowRFC3339Nano(),
	})
	w.WriteHeader(http.StatusNoContent)
}
