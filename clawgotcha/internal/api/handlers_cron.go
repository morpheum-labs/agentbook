package api

import (
	"encoding/json"
	"errors"
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

type createCronBody struct {
	Name           string `json:"name"`
	AgentName      string `json:"agent_name"`
	Schedule       string `json:"schedule"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	Prompt         string `json:"prompt"`
	Active         *bool  `json:"active"`
}

type patchCronBody struct {
	Name           *string `json:"name"`
	AgentName      *string `json:"agent_name"`
	Schedule       *string `json:"schedule"`
	TimeoutSeconds *int    `json:"timeout_seconds"`
	Prompt         *string `json:"prompt"`
	Active         *bool   `json:"active"`
}

func (s *Server) listCronJobs(w http.ResponseWriter, r *http.Request) {
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
	var out []db.SwarmCronJob
	if err := tx.Find(&out).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	if out == nil {
		out = []db.SwarmCronJob{}
	}
	sum, err := db.LoadRevisionSummary(s.db)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"cron_jobs":        out,
		"revision_summary": sum,
	})
}

func (s *Server) createCronJob(w http.ResponseWriter, r *http.Request) {
	var b createCronBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid json body", err))
		return
	}
	if strings.TrimSpace(b.Name) == "" {
		httperr.Write(w, r, httperr.BadRequest("name required", nil))
		return
	}
	agentName := strings.TrimSpace(b.AgentName)
	if agentName == "" {
		httperr.Write(w, r, httperr.BadRequest("agent_name required", nil))
		return
	}
	if err := agentExists(s.db, agentName); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httperr.Write(w, r, httperr.BadRequest("agent_name: no swarm agent with that name", err))
			return
		}
		httperr.Write(w, r, err)
		return
	}
	active := true
	if b.Active != nil {
		active = *b.Active
	}
	cj := db.SwarmCronJob{
		Name:           strings.TrimSpace(b.Name),
		AgentName:      agentName,
		Schedule:       strings.TrimSpace(b.Schedule),
		TimeoutSeconds: b.TimeoutSeconds,
		Prompt:         b.Prompt,
		Active:         active,
	}
	db.TouchCronRevision(&cj)
	if err := s.db.Create(&cj).Error; err != nil {
		if isUniqueViolation(err) {
			httperr.Write(w, r, httperr.BadRequest("duplicate", err))
			return
		}
		httperr.Write(w, r, err)
		return
	}
	s.emit(events.ChangeEvent{
		EventType:          events.EventCronUpdated,
		AffectedEntityType: events.EntityCronJob,
		AffectedIDs:        []string{cj.ID.String()},
		NewRevision:        cj.CurrentRevision,
		TS:                 events.NowRFC3339Nano(),
	})
	sum, err := db.LoadRevisionSummary(s.db)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"cron_job":         cj,
		"revision_summary": sum,
	})
}

func (s *Server) getCronJob(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid id", err))
		return
	}
	var cj db.SwarmCronJob
	if err := s.db.First(&cj, "id = ?", id).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	sum, err := db.LoadRevisionSummary(s.db)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"cron_job":         cj,
		"revision_summary": sum,
	})
}

func (s *Server) putCronJob(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid id", err))
		return
	}
	var b createCronBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid json body", err))
		return
	}
	if strings.TrimSpace(b.Name) == "" {
		httperr.Write(w, r, httperr.BadRequest("name required", nil))
		return
	}
	agentName := strings.TrimSpace(b.AgentName)
	if agentName == "" {
		httperr.Write(w, r, httperr.BadRequest("agent_name required", nil))
		return
	}
	if err := agentExists(s.db, agentName); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httperr.Write(w, r, httperr.BadRequest("agent_name: no swarm agent with that name", err))
			return
		}
		httperr.Write(w, r, err)
		return
	}
	var existing db.SwarmCronJob
	if err := s.db.First(&existing, "id = ?", id).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	active := existing.Active
	if b.Active != nil {
		active = *b.Active
	}
	cj := db.SwarmCronJob{
		ID:              id,
		CreatedAt:       existing.CreatedAt,
		Name:            strings.TrimSpace(b.Name),
		AgentName:       agentName,
		Schedule:        strings.TrimSpace(b.Schedule),
		TimeoutSeconds:  b.TimeoutSeconds,
		Prompt:          b.Prompt,
		Active:          active,
		CurrentRevision: existing.CurrentRevision,
	}
	var out db.SwarmCronJob
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&cj).Error; err != nil {
			return err
		}
		if err := db.IncrementCronRevision(tx, cj.ID); err != nil {
			return err
		}
		return tx.First(&out, "id = ?", cj.ID).Error
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
		EventType:          events.EventCronUpdated,
		AffectedEntityType: events.EntityCronJob,
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
		"cron_job":         out,
		"revision_summary": sum,
	})
}

func (s *Server) patchCronJob(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid id", err))
		return
	}
	var b patchCronBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid json body", err))
		return
	}
	var cj db.SwarmCronJob
	if err := s.db.First(&cj, "id = ?", id).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	updates := map[string]any{}
	if b.Name != nil {
		updates["name"] = strings.TrimSpace(*b.Name)
	}
	if b.Schedule != nil {
		updates["schedule"] = strings.TrimSpace(*b.Schedule)
	}
	if b.TimeoutSeconds != nil {
		updates["timeout_seconds"] = *b.TimeoutSeconds
	}
	if b.Prompt != nil {
		updates["prompt"] = *b.Prompt
	}
	if b.Active != nil {
		updates["active"] = *b.Active
	}
	if b.AgentName != nil {
		an := strings.TrimSpace(*b.AgentName)
		if an == "" {
			httperr.Write(w, r, httperr.BadRequest("agent_name cannot be empty", nil))
			return
		}
		if err := agentExists(s.db, an); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				httperr.Write(w, r, httperr.BadRequest("agent_name: no swarm agent with that name", err))
				return
			}
			httperr.Write(w, r, err)
			return
		}
		updates["agent_name"] = an
	}
	if len(updates) == 0 {
		sum, err := db.LoadRevisionSummary(s.db)
		if err != nil {
			httperr.Write(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"cron_job":         cj,
			"revision_summary": sum,
		})
		return
	}
	now := time.Now().UTC()
	updates["current_revision"] = gorm.Expr("current_revision + 1")
	updates["last_changed_at"] = now
	updates["updated_at"] = now
	if err := s.db.Model(&cj).Updates(updates).Error; err != nil {
		if isUniqueViolation(err) {
			httperr.Write(w, r, httperr.BadRequest("duplicate name", err))
			return
		}
		httperr.Write(w, r, err)
		return
	}
	if err := s.db.First(&cj, "id = ?", id).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	s.emit(events.ChangeEvent{
		EventType:          events.EventCronUpdated,
		AffectedEntityType: events.EntityCronJob,
		AffectedIDs:        []string{cj.ID.String()},
		NewRevision:        cj.CurrentRevision,
		TS:                 events.NowRFC3339Nano(),
	})
	sum, err := db.LoadRevisionSummary(s.db)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"cron_job":         cj,
		"revision_summary": sum,
	})
}

func (s *Server) deleteCronJob(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid id", err))
		return
	}
	var cj db.SwarmCronJob
	if err := s.db.First(&cj, "id = ?", id).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	now := time.Now().UTC()
	if err := s.db.Model(&db.SwarmCronJob{}).Unscoped().Where("id = ?", id).Updates(map[string]any{
		"deleted_at":       now,
		"current_revision": gorm.Expr("current_revision + 1"),
		"last_changed_at":  now,
	}).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	if err := s.db.Unscoped().First(&cj, "id = ?", id).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	s.emit(events.ChangeEvent{
		EventType:          events.EventCronDeleted,
		AffectedEntityType: events.EntityCronJob,
		AffectedIDs:        []string{cj.ID.String()},
		NewRevision:        cj.CurrentRevision,
		TS:                 events.NowRFC3339Nano(),
	})
	w.WriteHeader(http.StatusNoContent)
}
