package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/db"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/events"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/httperr"
	"gorm.io/gorm"
)

func defaultWebhookEventTypes() []string {
	return []string{
		events.EventAgentUpdated,
		events.EventAgentDeleted,
		events.EventCronUpdated,
		events.EventCronDeleted,
		events.EventConfigUpdated,
	}
}

func randomSecretHex(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

type registerInstanceBody struct {
	InstanceName string          `json:"instance_name"`
	InstanceType string          `json:"instance_type"`
	Version      string          `json:"version"`
	Hostname     string          `json:"hostname"`
	PublicURL    *string         `json:"public_url"`
	CallbackURL  string          `json:"callback_url"`
	Capabilities []string        `json:"capabilities"`
	Metadata     json.RawMessage `json:"metadata"`
	StartedAt    *time.Time      `json:"started_at"`
}

func (s *Server) registerInstance(w http.ResponseWriter, r *http.Request) {
	var b registerInstanceBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid json body", err))
		return
	}
	name := strings.TrimSpace(b.InstanceName)
	if name == "" || strings.TrimSpace(b.CallbackURL) == "" || strings.TrimSpace(b.Hostname) == "" || strings.TrimSpace(b.Version) == "" {
		httperr.Write(w, r, httperr.BadRequest("instance_name, callback_url, hostname, and version are required", nil))
		return
	}
	instType := strings.TrimSpace(b.InstanceType)
	if instType == "" {
		instType = "miroclaw"
	}
	caps := b.Capabilities
	if caps == nil {
		caps = []string{}
	}
	started := time.Now().UTC()
	if b.StartedAt != nil {
		started = b.StartedAt.UTC()
	}
	now := time.Now().UTC()

	var inst db.SwarmRuntimeInstance
	err := s.db.Where("instance_name = ?", name).First(&inst).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		inst = db.SwarmRuntimeInstance{
			InstanceName:    name,
			InstanceType:    instType,
			Version:         strings.TrimSpace(b.Version),
			Hostname:        strings.TrimSpace(b.Hostname),
			PublicURL:       b.PublicURL,
			CallbackURL:     strings.TrimSpace(b.CallbackURL),
			Capabilities:    caps,
			LastHeartbeatAt: &now,
			Status:          db.RuntimeStatusOnline,
			StartedAt:       started,
			Metadata:        b.Metadata,
		}
		if err := s.db.Create(&inst).Error; err != nil {
			httperr.Write(w, r, err)
			return
		}
	} else if err != nil {
		httperr.Write(w, r, err)
		return
	} else {
		inst.InstanceType = instType
		inst.Version = strings.TrimSpace(b.Version)
		inst.Hostname = strings.TrimSpace(b.Hostname)
		inst.PublicURL = b.PublicURL
		inst.CallbackURL = strings.TrimSpace(b.CallbackURL)
		inst.Capabilities = caps
		inst.LastHeartbeatAt = &now
		inst.Status = db.RuntimeStatusOnline
		inst.StartedAt = started
		if len(b.Metadata) > 0 {
			inst.Metadata = b.Metadata
		}
		if err := s.db.Save(&inst).Error; err != nil {
			httperr.Write(w, r, err)
			return
		}
	}

	var subCount int64
	if err := s.db.Model(&db.SwarmWebhookSubscription{}).Where("runtime_instance_id = ?", inst.ID).Count(&subCount).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	if subCount == 0 {
		sub := db.SwarmWebhookSubscription{
			RuntimeInstanceID: inst.ID,
			EventTypes:        defaultWebhookEventTypes(),
			Secret:            randomSecretHex(32),
			Enabled:           true,
		}
		if err := s.db.Create(&sub).Error; err != nil {
			httperr.Write(w, r, err)
			return
		}
	}

	sum, err := db.LoadRevisionSummary(s.db)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"instance":         inst,
		"revision_summary": sum,
	})
}

type heartbeatBody struct {
	Status   *string         `json:"status"`
	Metadata json.RawMessage `json:"metadata"`
}

func (s *Server) heartbeatInstance(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(chi.URLParam(r, "instance_name"))
	if name == "" {
		httperr.Write(w, r, httperr.BadRequest("instance_name required", nil))
		return
	}
	var b heartbeatBody
	_ = json.NewDecoder(r.Body).Decode(&b)

	var inst db.SwarmRuntimeInstance
	if err := s.db.Where("instance_name = ?", name).First(&inst).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	now := time.Now().UTC()
	inst.LastHeartbeatAt = &now
	if b.Status != nil && strings.TrimSpace(*b.Status) != "" {
		inst.Status = strings.TrimSpace(*b.Status)
	}
	if len(b.Metadata) > 0 {
		inst.Metadata = b.Metadata
	}
	if err := s.db.Save(&inst).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	if err := db.MarkStaleRuntimeInstancesOffline(s.db); err != nil {
		httperr.Write(w, r, err)
		return
	}
	sum, err := db.LoadRevisionSummary(s.db)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"revision_summary": sum})
}

func (s *Server) listInstances(w http.ResponseWriter, r *http.Request) {
	if err := db.MarkStaleRuntimeInstancesOffline(s.db); err != nil {
		httperr.Write(w, r, err)
		return
	}
	q := s.db.Model(&db.SwarmRuntimeInstance{})
	if st := strings.TrimSpace(r.URL.Query().Get("status")); st != "" {
		q = q.Where("status = ?", st)
	}
	var out []db.SwarmRuntimeInstance
	if err := q.Order("instance_name").Find(&out).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	if out == nil {
		out = []db.SwarmRuntimeInstance{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"instances": out})
}

func (s *Server) getInstance(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(chi.URLParam(r, "instance_name"))
	if name == "" {
		httperr.Write(w, r, httperr.BadRequest("instance_name required", nil))
		return
	}
	if err := db.MarkStaleRuntimeInstancesOffline(s.db); err != nil {
		httperr.Write(w, r, err)
		return
	}
	var inst db.SwarmRuntimeInstance
	if err := s.db.Where("instance_name = ?", name).First(&inst).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, inst)
}

func (s *Server) deleteInstance(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(chi.URLParam(r, "instance_name"))
	if name == "" {
		httperr.Write(w, r, httperr.BadRequest("instance_name required", nil))
		return
	}
	var inst db.SwarmRuntimeInstance
	if err := s.db.Where("instance_name = ?", name).First(&inst).Error; err != nil {
		httperr.Write(w, r, err)
		return
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&inst).Updates(map[string]interface{}{
			"status": db.RuntimeStatusOffline,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&db.SwarmWebhookSubscription{}).Where("runtime_instance_id = ?", inst.ID).Update("enabled", false).Error
	}); err != nil {
		httperr.Write(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
