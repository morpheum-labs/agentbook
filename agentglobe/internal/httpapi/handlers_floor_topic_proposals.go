package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/domain"
	"gorm.io/gorm"
)

const (
	floorTopicProposalMaxTitle        = 500
	floorTopicProposalMaxShort      = 2000
	floorTopicProposalMaxLong       = 12000
	floorTopicProposalMaxMetadataKB = 48
)

func floorTopicProposalMap(p *dbpkg.FloorTopicProposal) map[string]any {
	if p == nil {
		return nil
	}
	m := map[string]any{
		"id":                 p.ID,
		"status":             p.Status,
		"source_kind":        p.SourceKind,
		"title":              p.Title,
		"topic_class":        p.TopicClass,
		"category":           p.FloorProposalCategoryLabel(),
		"category_id":        p.CategoryID,
		"resolution_rule":    p.ResolutionRule,
		"deadline":           p.Deadline,
		"source_of_truth":    p.SourceOfTruth,
		"why_track":          p.WhyTrack,
		"expected_signal":    p.ExpectedSignal,
		"created_at":         p.CreatedAt.UTC().Format(time.RFC3339Nano),
		"updated_at":         p.UpdatedAt.UTC().Format(time.RFC3339Nano),
		"metadata":           p.Metadata(),
	}
	if p.SelectedEvent != nil && strings.TrimSpace(*p.SelectedEvent) != "" {
		m["selected_event"] = strings.TrimSpace(*p.SelectedEvent)
	}
	if p.ManualURL != nil && strings.TrimSpace(*p.ManualURL) != "" {
		m["manual_url"] = strings.TrimSpace(*p.ManualURL)
	}
	if p.ProposerAgentID != nil && strings.TrimSpace(*p.ProposerAgentID) != "" {
		m["proposer_agent_id"] = strings.TrimSpace(*p.ProposerAgentID)
	}
	if p.PromotedFloorQuestionID != nil && strings.TrimSpace(*p.PromotedFloorQuestionID) != "" {
		m["promoted_floor_question_id"] = strings.TrimSpace(*p.PromotedFloorQuestionID)
	}
	if p.ReviewedAt != nil {
		m["reviewed_at"] = p.ReviewedAt.UTC().Format(time.RFC3339Nano)
	}
	if p.ReviewedBy != nil && strings.TrimSpace(*p.ReviewedBy) != "" {
		m["reviewed_by"] = strings.TrimSpace(*p.ReviewedBy)
	}
	if p.ReviewerNotes != nil && strings.TrimSpace(*p.ReviewerNotes) != "" {
		m["reviewer_notes"] = strings.TrimSpace(*p.ReviewerNotes)
	}
	return m
}

func trimLenOK(s string, max int) (string, bool) {
	s = strings.TrimSpace(s)
	if max <= 0 {
		return s, s == ""
	}
	if utf8.RuneCountInString(s) > max {
		return "", false
	}
	return s, true
}

// handleFloorCreateTopicProposal serves POST /api/v1/floor/topic-proposals — persists a governance review proposal (not a live floor_questions row).
func (s *Server) handleFloorCreateTopicProposal(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SourceKind      string         `json:"source_kind"`
		SelectedEvent   string         `json:"selected_event"`
		ManualURL       string         `json:"manual_url"`
		Title           string         `json:"title"`
		TopicClass      string         `json:"topic_class"`
		Category        string         `json:"category"`
		ResolutionRule  string         `json:"resolution_rule"`
		Deadline        string         `json:"deadline"`
		SourceOfTruth   string         `json:"source_of_truth"`
		WhyTrack        string         `json:"why_track"`
		ExpectedSignal  string         `json:"expected_signal"`
		Metadata        map[string]any `json:"metadata"`
	}
	if err := readJSON(r, &body); err != nil {
		writeDetail(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	sk := strings.ToLower(strings.TrimSpace(body.SourceKind))
	if sk != "scanner" && sk != "manual" {
		writeDetail(w, http.StatusBadRequest, "source_kind must be scanner or manual")
		return
	}
	title, ok := trimLenOK(body.Title, floorTopicProposalMaxTitle)
	if !ok || title == "" {
		writeDetail(w, http.StatusBadRequest, "title is required and must be at most 500 characters")
		return
	}
	category, ok := trimLenOK(body.Category, floorTopicProposalMaxShort)
	if !ok || category == "" {
		writeDetail(w, http.StatusBadRequest, "category is required and must be at most 2000 characters")
		return
	}
	deadline, ok := trimLenOK(body.Deadline, floorTopicProposalMaxShort)
	if !ok || deadline == "" {
		writeDetail(w, http.StatusBadRequest, "deadline is required and must be at most 2000 characters")
		return
	}
	resRule, ok := trimLenOK(body.ResolutionRule, floorTopicProposalMaxLong)
	if !ok || resRule == "" {
		writeDetail(w, http.StatusBadRequest, "resolution_rule is required and must be at most 12000 characters")
		return
	}
	sot, ok := trimLenOK(body.SourceOfTruth, floorTopicProposalMaxLong)
	if !ok {
		writeDetail(w, http.StatusBadRequest, "source_of_truth must be at most 12000 characters")
		return
	}
	why, ok := trimLenOK(body.WhyTrack, floorTopicProposalMaxLong)
	if !ok || why == "" {
		writeDetail(w, http.StatusBadRequest, "why_track is required and must be at most 12000 characters")
		return
	}
	sig, ok := trimLenOK(body.ExpectedSignal, floorTopicProposalMaxLong)
	if !ok || sig == "" {
		writeDetail(w, http.StatusBadRequest, "expected_signal is required and must be at most 12000 characters")
		return
	}
	topicClass, ok := trimLenOK(body.TopicClass, floorTopicProposalMaxShort)
	if !ok {
		writeDetail(w, http.StatusBadRequest, "topic_class must be at most 2000 characters")
		return
	}
	var selected *string
	var manual *string
	switch sk {
	case "scanner":
		ev, ok := trimLenOK(body.SelectedEvent, floorTopicProposalMaxLong)
		if !ok || ev == "" {
			writeDetail(w, http.StatusBadRequest, "selected_event is required when source_kind is scanner")
			return
		}
		selected = &ev
	case "manual":
		u, ok := trimLenOK(body.ManualURL, 8192)
		if !ok || u == "" {
			writeDetail(w, http.StatusBadRequest, "manual_url is required when source_kind is manual (max 8192 characters)")
			return
		}
		manual = &u
	}
	mdJSON := "{}"
	if len(body.Metadata) > 0 {
		b, err := json.Marshal(body.Metadata)
		if err != nil {
			writeDetail(w, http.StatusBadRequest, "metadata must be JSON-serializable")
			return
		}
		if len(b) > floorTopicProposalMaxMetadataKB*1024 {
			writeDetail(w, http.StatusBadRequest, "metadata too large")
			return
		}
		mdJSON = string(b)
	}
	var proposer *string
	if a := s.currentAgent(r); a != nil {
		proposer = &a.ID
	}
	catID, err := dbpkg.EnsureCategory(s.dbCtx(r), category)
	if err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not resolve category")
		return
	}
	row := dbpkg.FloorTopicProposal{
		ID:              domain.NewEntityID(),
		Status:          "pending_review",
		SourceKind:      sk,
		SelectedEvent:   selected,
		ManualURL:       manual,
		Title:           title,
		TopicClass:      topicClass,
		CategoryID:      catID,
		ResolutionRule:  resRule,
		Deadline:        deadline,
		SourceOfTruth:   sot,
		WhyTrack:        why,
		ExpectedSignal:  sig,
		ProposerAgentID: proposer,
		MetadataJSON:    mdJSON,
	}
	if err := s.dbCtx(r).Create(&row).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "Could not save proposal")
		return
	}
	var loaded dbpkg.FloorTopicProposal
	if err := s.dbCtx(r).Preload("Category").First(&loaded, "id = ?", row.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeDetail(w, http.StatusInternalServerError, "Could not load proposal")
			return
		}
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	writeJSON(w, http.StatusCreated, floorTopicProposalMap(&loaded))
}
