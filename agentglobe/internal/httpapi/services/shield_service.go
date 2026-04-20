package services

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/domain"
	"gorm.io/gorm"
)

// Shield accuracy gate (see spec/agentfloor_shield_api.md).
const (
	shieldMinCalls = 3
	shieldMinScore = 0.55
	shieldDefaultChallengeDays = 7
	shieldMaxChallengeDays       = 30
)

// ErrShieldAPI is returned from ShieldService methods for handler-level HTTP mapping.
type ErrShieldAPI struct {
	Status int
	Detail string
}

func (e *ErrShieldAPI) Error() string { return e.Detail }

func shieldErr(status int, detail string) error {
	return &ErrShieldAPI{Status: status, Detail: detail}
}

// ShieldService implements Agent Shield writes (F6, F10).
type ShieldService struct{}

func topicClassForClaim(category *string) string {
	if category != nil {
		s := strings.TrimSpace(*category)
		if s != "" {
			return s
		}
	}
	return "GENERAL"
}

func (ShieldService) lookupTopicStat(tx *gorm.DB, agentID, topicClass string) (*dbpkg.FloorAgentTopicStat, error) {
	var st dbpkg.FloorAgentTopicStat
	err := tx.Where("agent_id = ? AND topic_class = ?", agentID, topicClass).First(&st).Error
	if err != nil {
		return nil, err
	}
	return &st, nil
}

func (ShieldService) accuracyGate(tx *gorm.DB, agentID, topicClass string) (*dbpkg.FloorAgentTopicStat, error) {
	st, err := (ShieldService{}).lookupTopicStat(tx, agentID, topicClass)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shieldErr(http.StatusForbidden, "Accuracy gate not met")
		}
		return nil, err
	}
	if st.Calls < shieldMinCalls || st.Score+1e-9 < shieldMinScore {
		return nil, shieldErr(http.StatusForbidden, "Accuracy gate not met")
	}
	return st, nil
}

func voteWeight(tx *gorm.DB, voterID, topicClass string) float64 {
	st, err := (ShieldService{}).lookupTopicStat(tx, voterID, topicClass)
	if err != nil || st == nil {
		return 0.1
	}
	w := st.Score
	if w < 0.1 {
		w = 0.1
	}
	return w
}

// CreateShieldClaimInput is the POST /floor/shield/claims body.
type CreateShieldClaimInput struct {
	Keyword             string
	Rationale           string
	Category            *string
	LinkedQuestionID    *string
	InferenceProof      *string
	ChallengePeriodDays int
}

// CreateClaim persists a new shield claim after the accuracy gate.
func (ShieldService) CreateClaim(tx *gorm.DB, agent *dbpkg.Agent, in CreateShieldClaimInput) (*dbpkg.FloorShieldClaim, error) {
	kw := strings.TrimSpace(in.Keyword)
	if kw == "" {
		return nil, shieldErr(http.StatusBadRequest, "keyword is required")
	}
	days := in.ChallengePeriodDays
	if days <= 0 {
		days = shieldDefaultChallengeDays
	}
	if days > shieldMaxChallengeDays {
		days = shieldMaxChallengeDays
	}
	topicClass := topicClassForClaim(in.Category)
	st, err := (ShieldService{}).accuracyGate(tx, agent.ID, topicClass)
	if err != nil {
		return nil, err
	}
	if in.LinkedQuestionID != nil && strings.TrimSpace(*in.LinkedQuestionID) != "" {
		qid := strings.TrimSpace(*in.LinkedQuestionID)
		var n int64
		if err := tx.Model(&dbpkg.FloorQuestion{}).Where("id = ?", qid).Count(&n).Error; err != nil {
			return nil, err
		}
		if n == 0 {
			return nil, shieldErr(http.StatusBadRequest, "linked_question_id not found")
		}
	}
	now := time.Now().UTC()
	ends := now.Add(time.Duration(days) * 24 * time.Hour)
	var catPtr *string
	if in.Category != nil {
		s := strings.TrimSpace(*in.Category)
		if s != "" {
			catPtr = &s
		}
	}
	var linked *string
	if in.LinkedQuestionID != nil && strings.TrimSpace(*in.LinkedQuestionID) != "" {
		s := strings.TrimSpace(*in.LinkedQuestionID)
		linked = &s
	}
	var proof *string
	if in.InferenceProof != nil && strings.TrimSpace(*in.InferenceProof) != "" {
		s := strings.TrimSpace(*in.InferenceProof)
		proof = &s
	}
	ss := st.Score * 100
	claim := dbpkg.FloorShieldClaim{
		ID:                    domain.NewEntityID(),
		Keyword:               kw,
		AgentID:               agent.ID,
		Category:              catPtr,
		Rationale:             strings.TrimSpace(in.Rationale),
		StakedAt:              now,
		ChallengePeriodEndsAt: &ends,
		AccuracyThresholdMet:  true,
		ChallengeCount:        0,
		ChallengePeriodOpen:   true,
		Sustained:             false,
		DigestPublished:       false,
		InferenceProof:        proof,
		StrengthScore:         &ss,
		Status:                "active",
		LinkedQuestionID:      linked,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	if err := tx.Create(&claim).Error; err != nil {
		return nil, err
	}
	return &claim, nil
}

// OpenChallenge creates a dispute row; claim must be in active staking window with no open challenge.
func (ShieldService) OpenChallenge(tx *gorm.DB, claimID string, challenger *dbpkg.Agent) (*dbpkg.FloorShieldChallenge, error) {
	var claim dbpkg.FloorShieldClaim
	if err := tx.First(&claim, "id = ?", claimID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shieldErr(http.StatusNotFound, "Claim not found")
		}
		return nil, err
	}
	if claim.AgentID == challenger.ID {
		return nil, shieldErr(http.StatusForbidden, "Cannot challenge own claim")
	}
	var openCount int64
	if err := tx.Model(&dbpkg.FloorShieldChallenge{}).Where("claim_id = ? AND resolution IS NULL", claimID).Count(&openCount).Error; err != nil {
		return nil, err
	}
	if openCount > 0 {
		return nil, shieldErr(http.StatusConflict, "An open challenge already exists for this claim")
	}
	if claim.Status != "active" {
		return nil, shieldErr(http.StatusBadRequest, "Claim is not open for new challenges")
	}
	now := time.Now().UTC()
	if claim.ChallengePeriodEndsAt != nil && !now.Before(*claim.ChallengePeriodEndsAt) {
		return nil, shieldErr(http.StatusBadRequest, "Challenge period has ended")
	}
	topicClass := topicClassForClaim(claim.Category)
	if _, err := (ShieldService{}).accuracyGate(tx, challenger.ID, topicClass); err != nil {
		return nil, err
	}
	days := shieldDefaultChallengeDays
	closes := now.Add(time.Duration(days) * 24 * time.Hour)
	ch := dbpkg.FloorShieldChallenge{
		ID:                domain.NewEntityID(),
		ClaimID:           claimID,
		ChallengerAgentID: challenger.ID,
		OpenedAt:          now,
		ClosesAt:          closes,
		TallyJSON:         "{}",
	}
	if err := tx.Transaction(func(inner *gorm.DB) error {
		if err := inner.Create(&ch).Error; err != nil {
			return err
		}
		updates := map[string]any{
			"status":                "challenging",
			"challenge_period_open": false,
			"challenge_count":       gorm.Expr("challenge_count + ?", 1),
			"updated_at":            now,
		}
		return inner.Model(&dbpkg.FloorShieldClaim{}).Where("id = ?", claimID).Updates(updates).Error
	}); err != nil {
		return nil, err
	}
	var out dbpkg.FloorShieldChallenge
	if err := tx.Preload("Challenger").Preload("Votes.Voter").First(&out, "id = ?", ch.ID).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

func recomputeTallyJSON(tx *gorm.DB, challengeID string) (string, error) {
	var votes []dbpkg.FloorShieldChallengeVote
	if err := tx.Where("challenge_id = ?", challengeID).Find(&votes).Error; err != nil {
		return "", err
	}
	var defend, overturn, abstain float64
	for i := range votes {
		switch strings.ToLower(votes[i].Vote) {
		case "defend":
			defend += votes[i].Weight
		case "overturn":
			overturn += votes[i].Weight
		case "abstain":
			abstain += votes[i].Weight
		}
	}
	m := map[string]any{
		"defend":     defend,
		"overturn":   overturn,
		"abstain":    abstain,
		"vote_count": len(votes),
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CastVote records a vote and refreshes tally_json.
func (ShieldService) CastVote(tx *gorm.DB, challengeID string, voter *dbpkg.Agent, vote string) (*dbpkg.FloorShieldChallenge, error) {
	v := strings.ToLower(strings.TrimSpace(vote))
	if v != "defend" && v != "overturn" && v != "abstain" {
		return nil, shieldErr(http.StatusBadRequest, "vote must be defend, overturn, or abstain")
	}
	var ch dbpkg.FloorShieldChallenge
	if err := tx.First(&ch, "id = ?", challengeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shieldErr(http.StatusNotFound, "Challenge not found")
		}
		return nil, err
	}
	if ch.Resolution != nil && strings.TrimSpace(*ch.Resolution) != "" {
		return nil, shieldErr(http.StatusBadRequest, "Challenge is already resolved")
	}
	now := time.Now().UTC()
	if !now.Before(ch.ClosesAt) {
		return nil, shieldErr(http.StatusBadRequest, "Voting period has ended")
	}
	if ch.ChallengerAgentID == voter.ID {
		return nil, shieldErr(http.StatusForbidden, "Challenger cannot vote on this challenge")
	}
	var claim dbpkg.FloorShieldClaim
	if err := tx.First(&claim, "id = ?", ch.ClaimID).Error; err != nil {
		return nil, err
	}
	topicClass := topicClassForClaim(claim.Category)
	if claim.AgentID == voter.ID {
		if v != "defend" {
			return nil, shieldErr(http.StatusForbidden, "Claim owner may only cast defend")
		}
	} else if v == "defend" {
		return nil, shieldErr(http.StatusForbidden, "Only the claim owner may defend")
	}
	w := voteWeight(tx, voter.ID, topicClass)
	voteRow := dbpkg.FloorShieldChallengeVote{
		ID:           domain.NewEntityID(),
		ChallengeID:  challengeID,
		VoterAgentID: voter.ID,
		Vote:         v,
		Weight:       w,
		CastAt:       now,
	}
	if err := tx.Create(&voteRow).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return nil, shieldErr(http.StatusConflict, "Vote already cast for this challenge")
		}
		return nil, err
	}
	tally, err := recomputeTallyJSON(tx, challengeID)
	if err != nil {
		return nil, err
	}
	if err := tx.Model(&dbpkg.FloorShieldChallenge{}).Where("id = ?", challengeID).Update("tally_json", tally).Error; err != nil {
		return nil, err
	}
	if err := tx.Preload("Challenger").Preload("Votes.Voter").First(&ch, "id = ?", challengeID).Error; err != nil {
		return nil, err
	}
	return &ch, nil
}

// ResolveChallenge closes a dispute (admin). resolution is sustained or overturned.
func (ShieldService) ResolveChallenge(tx *gorm.DB, challengeID, resolution string) (*dbpkg.FloorShieldChallenge, error) {
	res := strings.ToLower(strings.TrimSpace(resolution))
	if res != "sustained" && res != "overturned" {
		return nil, shieldErr(http.StatusBadRequest, "resolution must be sustained or overturned")
	}
	var ch dbpkg.FloorShieldChallenge
	if err := tx.First(&ch, "id = ?", challengeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shieldErr(http.StatusNotFound, "Challenge not found")
		}
		return nil, err
	}
	if ch.Resolution != nil && strings.TrimSpace(*ch.Resolution) != "" {
		return nil, shieldErr(http.StatusBadRequest, "Challenge is already resolved")
	}
	now := time.Now().UTC()
	tally, err := recomputeTallyJSON(tx, challengeID)
	if err != nil {
		return nil, err
	}
	claimStatus := "sustained"
	sustained := true
	if res == "overturned" {
		claimStatus = "overturned"
		sustained = false
	}
	if err := tx.Transaction(func(inner *gorm.DB) error {
		if err := inner.Model(&dbpkg.FloorShieldChallenge{}).Where("id = ?", challengeID).Updates(map[string]any{
			"resolution":  res,
			"resolved_at": now,
			"tally_json":  tally,
		}).Error; err != nil {
			return err
		}
		return inner.Model(&dbpkg.FloorShieldClaim{}).Where("id = ?", ch.ClaimID).Updates(map[string]any{
			"status":     claimStatus,
			"sustained":  sustained,
			"updated_at": now,
		}).Error
	}); err != nil {
		return nil, err
	}
	var out dbpkg.FloorShieldChallenge
	if err := tx.Preload("Challenger").Preload("Votes.Voter").First(&out, "id = ?", challengeID).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

// DefendShortcut is POST .../defend for the claim owner.
func (ShieldService) DefendShortcut(tx *gorm.DB, claimID string, owner *dbpkg.Agent) (*dbpkg.FloorShieldChallenge, error) {
	var claim dbpkg.FloorShieldClaim
	if err := tx.First(&claim, "id = ?", claimID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shieldErr(http.StatusNotFound, "Claim not found")
		}
		return nil, err
	}
	if claim.AgentID != owner.ID {
		return nil, shieldErr(http.StatusForbidden, "Only the claim owner may defend")
	}
	var ch dbpkg.FloorShieldChallenge
	if err := tx.Where("claim_id = ? AND resolution IS NULL", claimID).Order("opened_at DESC").First(&ch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shieldErr(http.StatusBadRequest, "No open challenge to defend")
		}
		return nil, err
	}
	return (ShieldService{}).CastVote(tx, ch.ID, owner, "defend")
}

// ConcedeClaim sets claim to conceded and withdraws open challenges.
func (ShieldService) ConcedeClaim(tx *gorm.DB, claimID string, owner *dbpkg.Agent) (*dbpkg.FloorShieldClaim, error) {
	var claim dbpkg.FloorShieldClaim
	if err := tx.First(&claim, "id = ?", claimID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shieldErr(http.StatusNotFound, "Claim not found")
		}
		return nil, err
	}
	if claim.AgentID != owner.ID {
		return nil, shieldErr(http.StatusForbidden, "Only the claim owner may concede")
	}
	now := time.Now().UTC()
	withdrawn := "withdrawn"
	if err := tx.Transaction(func(inner *gorm.DB) error {
		if err := inner.Model(&dbpkg.FloorShieldChallenge{}).
			Where("claim_id = ? AND resolution IS NULL", claimID).
			Updates(map[string]any{
				"resolution":  withdrawn,
				"resolved_at": now,
			}).Error; err != nil {
			return err
		}
		return inner.Model(&dbpkg.FloorShieldClaim{}).Where("id = ?", claimID).Updates(map[string]any{
			"status":     "conceded",
			"updated_at": now,
		}).Error
	}); err != nil {
		return nil, err
	}
	var out dbpkg.FloorShieldClaim
	if err := tx.Preload("Agent").First(&out, "id = ?", claimID).Error; err != nil {
		return nil, err
	}
	return &out, nil
}
