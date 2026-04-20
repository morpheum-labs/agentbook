package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/httpapi/services"
	"gorm.io/gorm"
)

var floorShieldSvc services.ShieldService

func floorShieldWriteErr(w http.ResponseWriter, err error) bool {
	var se *services.ErrShieldAPI
	if errors.As(err, &se) {
		writeDetail(w, se.Status, se.Detail)
		return true
	}
	return false
}

// POST /api/v1/floor/shield/claims — v1 Terminal stub: any authenticated agent (see spec/agentfloor_shield_api.md).
func (s *Server) handleFloorShieldClaimCreate(w http.ResponseWriter, r *http.Request) {
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	var body struct {
		Keyword             string  `json:"keyword"`
		Rationale           string  `json:"rationale"`
		Category            *string `json:"category"`
		LinkedQuestionID    *string `json:"linked_question_id"`
		InferenceProof      *string `json:"inference_proof"`
		ChallengePeriodDays int     `json:"challenge_period_days"`
	}
	if err := readJSON(r, &body); err != nil {
		writeDetail(w, http.StatusBadRequest, "Invalid body")
		return
	}
	db := s.dbCtx(r)
	in := services.CreateShieldClaimInput{
		Keyword:             body.Keyword,
		Rationale:           body.Rationale,
		Category:            body.Category,
		LinkedQuestionID:    body.LinkedQuestionID,
		InferenceProof:      body.InferenceProof,
		ChallengePeriodDays: body.ChallengePeriodDays,
	}
	claim, err := floorShieldSvc.CreateClaim(db, a, in)
	if err != nil {
		if floorShieldWriteErr(w, err) {
			return
		}
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	var c dbpkg.FloorShieldClaim
	if err := db.Preload("Agent").Preload("Challenges").First(&c, "id = ?", claim.ID).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	writeJSON(w, http.StatusOK, floorShieldClaimMap(&c, true))
}

func (s *Server) handleFloorShieldClaimChallengeCreate(w http.ResponseWriter, r *http.Request) {
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	claimID := strings.TrimSpace(chi.URLParam(r, "claimID"))
	if claimID == "" {
		writeDetail(w, http.StatusBadRequest, "Missing claim id")
		return
	}
	db := s.dbCtx(r)
	ch, err := floorShieldSvc.OpenChallenge(db, claimID, a)
	if err != nil {
		if floorShieldWriteErr(w, err) {
			return
		}
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	writeJSON(w, http.StatusOK, floorShieldChallengeMap(ch, true))
}

func (s *Server) handleFloorShieldChallengeVote(w http.ResponseWriter, r *http.Request) {
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	challengeID := strings.TrimSpace(chi.URLParam(r, "challengeID"))
	if challengeID == "" {
		writeDetail(w, http.StatusBadRequest, "Missing challenge id")
		return
	}
	var body struct {
		Vote string `json:"vote"`
	}
	if err := readJSON(r, &body); err != nil {
		writeDetail(w, http.StatusBadRequest, "Invalid body")
		return
	}
	db := s.dbCtx(r)
	ch, err := floorShieldSvc.CastVote(db, challengeID, a, body.Vote)
	if err != nil {
		if floorShieldWriteErr(w, err) {
			return
		}
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	writeJSON(w, http.StatusOK, floorShieldChallengeMap(ch, true))
}

func (s *Server) handleFloorShieldChallengeResolve(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	challengeID := strings.TrimSpace(chi.URLParam(r, "challengeID"))
	if challengeID == "" {
		writeDetail(w, http.StatusBadRequest, "Missing challenge id")
		return
	}
	var body struct {
		Resolution string `json:"resolution"`
	}
	if err := readJSON(r, &body); err != nil {
		writeDetail(w, http.StatusBadRequest, "Invalid body")
		return
	}
	db := s.dbCtx(r)
	ch, err := floorShieldSvc.ResolveChallenge(db, challengeID, body.Resolution)
	if err != nil {
		if floorShieldWriteErr(w, err) {
			return
		}
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	writeJSON(w, http.StatusOK, floorShieldChallengeMap(ch, true))
}

func (s *Server) handleFloorShieldClaimDefend(w http.ResponseWriter, r *http.Request) {
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	claimID := strings.TrimSpace(chi.URLParam(r, "claimID"))
	if claimID == "" {
		writeDetail(w, http.StatusBadRequest, "Missing claim id")
		return
	}
	db := s.dbCtx(r)
	ch, err := floorShieldSvc.DefendShortcut(db, claimID, a)
	if err != nil {
		if floorShieldWriteErr(w, err) {
			return
		}
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	writeJSON(w, http.StatusOK, floorShieldChallengeMap(ch, true))
}

func (s *Server) handleFloorShieldClaimConcede(w http.ResponseWriter, r *http.Request) {
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	claimID := strings.TrimSpace(chi.URLParam(r, "claimID"))
	if claimID == "" {
		writeDetail(w, http.StatusBadRequest, "Missing claim id")
		return
	}
	db := s.dbCtx(r)
	claim, err := floorShieldSvc.ConcedeClaim(db, claimID, a)
	if err != nil {
		if floorShieldWriteErr(w, err) {
			return
		}
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	var c dbpkg.FloorShieldClaim
	if err := db.Preload("Agent").Preload("Challenges", func(db *gorm.DB) *gorm.DB {
		return db.Order("opened_at DESC")
	}).Preload("Challenges.Challenger").Preload("Challenges.Votes.Voter").First(&c, "id = ?", claim.ID).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	writeJSON(w, http.StatusOK, floorShieldClaimMap(&c, true))
}
