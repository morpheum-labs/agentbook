package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"gorm.io/gorm"
)

const floorMaxLimit = 50

func floorParsePagination(r *http.Request) (limit, offset int) {
	limit = floorMaxLimit
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
			if limit > floorMaxLimit {
				limit = floorMaxLimit
			}
		}
	}
	offset = 0
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	return limit, offset
}

func floorDecodeJSONObject(raw string) map[string]any {
	if strings.TrimSpace(raw) == "" || raw == "{}" {
		return map[string]any{}
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil || m == nil {
		return map[string]any{}
	}
	return m
}

func floorDecodeJSONArray(raw string) []any {
	if strings.TrimSpace(raw) == "" || raw == "[]" {
		return []any{}
	}
	var a []any
	if err := json.Unmarshal([]byte(raw), &a); err != nil || a == nil {
		return []any{}
	}
	return a
}

func floorQuestionMap(q *dbpkg.FloorQuestion) map[string]any {
	m := map[string]any{
		"id":                     q.ID,
		"title":                  q.Title,
		"category":               q.Category,
		"resolution_condition":   q.ResolutionCondition,
		"deadline":               q.Deadline,
		"probability":            q.Probability,
		"probability_delta":      q.ProbabilityDelta,
		"agent_count":            q.AgentCount,
		"staked_count":           q.StakedCount,
		"status":                 q.Status,
		"cluster_breakdown":      floorDecodeJSONObject(q.ClusterBreakdownJSON),
		"created_at":             q.CreatedAt.UTC().Format(time.RFC3339Nano),
		"updated_at":             q.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if q.ZkVerifiedPct != nil {
		m["zk_verified_pct"] = *q.ZkVerifiedPct
	} else {
		m["zk_verified_pct"] = nil
	}
	return m
}

func floorDigestMap(d *dbpkg.FloorDigestEntry) map[string]any {
	m := map[string]any{
		"id":                 d.ID,
		"question_id":        d.QuestionID,
		"digest_date":        d.DigestDate,
		"consensus_level":    d.ConsensusLevel,
		"probability":        d.Probability,
		"probability_delta":  d.ProbabilityDelta,
		"summary":            d.Summary,
		"cluster_breakdown":  floorDecodeJSONObject(d.ClusterBreakdownJSON),
		"created_at":         d.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
	if d.TopLongAgentID != nil {
		m["top_long_agent_id"] = *d.TopLongAgentID
	} else {
		m["top_long_agent_id"] = nil
	}
	if d.TopShortAgentID != nil {
		m["top_short_agent_id"] = *d.TopShortAgentID
	} else {
		m["top_short_agent_id"] = nil
	}
	if d.LlmIndexHits != nil {
		m["llm_index_hits"] = *d.LlmIndexHits
	} else {
		m["llm_index_hits"] = nil
	}
	return m
}

func floorPositionMap(p *dbpkg.FloorPosition) map[string]any {
	name := ""
	if p.Agent.ID != "" {
		name = p.Agent.Name
	}
	m := map[string]any{
		"id":                       p.ID,
		"question_id":              p.QuestionID,
		"agent_id":                 p.AgentID,
		"agent_name":               name,
		"direction":                p.Direction,
		"staked_at":                p.StakedAt.UTC().Format(time.RFC3339Nano),
		"body":                     p.Body,
		"language":                 p.Language,
		"resolved":                 p.Resolved,
		"outcome":                  p.Outcome,
		"challenge_open":           p.ChallengeOpen,
		"created_at":               p.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
	if p.AccuracyScoreAtStake != nil {
		m["accuracy_score_at_stake"] = *p.AccuracyScoreAtStake
	} else {
		m["accuracy_score_at_stake"] = nil
	}
	if p.InferenceProof != nil {
		m["inference_proof"] = *p.InferenceProof
	} else {
		m["inference_proof"] = nil
	}
	if p.ProofType != nil {
		m["proof_type"] = *p.ProofType
	} else {
		m["proof_type"] = nil
	}
	if p.RegionalCluster != nil {
		m["regional_cluster"] = *p.RegionalCluster
	} else {
		m["regional_cluster"] = nil
	}
	if p.SourcePostID != nil {
		m["source_post_id"] = *p.SourcePostID
	} else {
		m["source_post_id"] = nil
	}
	if p.SourceCommentID != nil {
		m["source_comment_id"] = *p.SourceCommentID
	} else {
		m["source_comment_id"] = nil
	}
	return m
}

func floorTopicStatMap(t *dbpkg.FloorAgentTopicStat) map[string]any {
	return map[string]any{
		"agent_id":     t.AgentID,
		"topic_class":  t.TopicClass,
		"calls":        t.Calls,
		"correct":      t.Correct,
		"score":        t.Score,
		"updated_at":   t.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func floorProbabilityPointMap(pt *dbpkg.FloorQuestionProbabilityPoint) map[string]any {
	return map[string]any{
		"id":           pt.ID,
		"question_id":  pt.QuestionID,
		"captured_at":  pt.CapturedAt.UTC().Format(time.RFC3339Nano),
		"probability":  pt.Probability,
		"source":       pt.Source,
	}
}

func floorShieldClaimMap(c *dbpkg.FloorShieldClaim, withChallenges bool) map[string]any {
	agentName := ""
	if c.Agent.ID != "" {
		agentName = c.Agent.Name
	}
	m := map[string]any{
		"id":                       c.ID,
		"keyword":                  c.Keyword,
		"agent_id":                 c.AgentID,
		"agent_name":               agentName,
		"rationale":                c.Rationale,
		"staked_at":                c.StakedAt.UTC().Format(time.RFC3339Nano),
		"accuracy_threshold_met":   c.AccuracyThresholdMet,
		"challenge_count":          c.ChallengeCount,
		"challenge_period_open":      c.ChallengePeriodOpen,
		"sustained":                c.Sustained,
		"digest_published":         c.DigestPublished,
		"status":                   c.Status,
		"created_at":               c.CreatedAt.UTC().Format(time.RFC3339Nano),
		"updated_at":               c.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if c.Category != nil {
		m["category"] = *c.Category
	} else {
		m["category"] = nil
	}
	if c.ChallengePeriodEndsAt != nil {
		m["challenge_period_ends_at"] = c.ChallengePeriodEndsAt.UTC().Format(time.RFC3339Nano)
	} else {
		m["challenge_period_ends_at"] = nil
	}
	if c.InferenceProof != nil {
		m["inference_proof"] = *c.InferenceProof
	} else {
		m["inference_proof"] = nil
	}
	if c.StrengthScore != nil {
		m["strength_score"] = *c.StrengthScore
	} else {
		m["strength_score"] = nil
	}
	if c.LinkedQuestionID != nil {
		m["linked_question_id"] = *c.LinkedQuestionID
	} else {
		m["linked_question_id"] = nil
	}
	if withChallenges {
		chs := make([]map[string]any, 0, len(c.Challenges))
		for i := range c.Challenges {
			chs = append(chs, floorShieldChallengeMap(&c.Challenges[i], true))
		}
		m["challenges"] = chs
	}
	return m
}

func floorShieldChallengeMap(c *dbpkg.FloorShieldChallenge, withVotes bool) map[string]any {
	chName := ""
	if c.Challenger.ID != "" {
		chName = c.Challenger.Name
	}
	m := map[string]any{
		"id":                    c.ID,
		"claim_id":              c.ClaimID,
		"challenger_agent_id":   c.ChallengerAgentID,
		"challenger_agent_name": chName,
		"opened_at":             c.OpenedAt.UTC().Format(time.RFC3339Nano),
		"closes_at":             c.ClosesAt.UTC().Format(time.RFC3339Nano),
		"tally":                 floorDecodeJSONObject(c.TallyJSON),
	}
	if c.Resolution != nil {
		m["resolution"] = *c.Resolution
	} else {
		m["resolution"] = nil
	}
	if c.ResolvedAt != nil {
		m["resolved_at"] = c.ResolvedAt.UTC().Format(time.RFC3339Nano)
	} else {
		m["resolved_at"] = nil
	}
	if withVotes {
		votes := make([]map[string]any, 0, len(c.Votes))
		for i := range c.Votes {
			votes = append(votes, floorShieldVoteMap(&c.Votes[i]))
		}
		m["votes"] = votes
	}
	return m
}

func floorShieldVoteMap(v *dbpkg.FloorShieldChallengeVote) map[string]any {
	vName := ""
	if v.Voter.ID != "" {
		vName = v.Voter.Name
	}
	return map[string]any{
		"id":              v.ID,
		"challenge_id":    v.ChallengeID,
		"voter_agent_id":  v.VoterAgentID,
		"voter_agent_name": vName,
		"vote":            v.Vote,
		"weight":          v.Weight,
		"cast_at":         v.CastAt.UTC().Format(time.RFC3339Nano),
	}
}

func floorPositionChallengeMap(c *dbpkg.FloorPositionChallenge) map[string]any {
	chName := ""
	if c.Challenger.ID != "" {
		chName = c.Challenger.Name
	}
	m := map[string]any{
		"id":                   c.ID,
		"position_id":          c.PositionID,
		"challenger_agent_id":  c.ChallengerAgentID,
		"challenger_agent_name": chName,
		"status":               c.Status,
		"opened_at":            c.OpenedAt.UTC().Format(time.RFC3339Nano),
	}
	if c.ResolvedAt != nil {
		m["resolved_at"] = c.ResolvedAt.UTC().Format(time.RFC3339Nano)
	} else {
		m["resolved_at"] = nil
	}
	if c.ResolutionNotes != nil {
		m["resolution_notes"] = *c.ResolutionNotes
	} else {
		m["resolution_notes"] = nil
	}
	return m
}

func floorArticleMap(a *dbpkg.FloorResearchArticle) map[string]any {
	m := map[string]any{
		"id":           a.ID,
		"title":        a.Title,
		"summary":      a.Summary,
		"cluster_tags": floorDecodeJSONArray(a.ClusterTagsJSON),
		"created_at":   a.CreatedAt.UTC().Format(time.RFC3339Nano),
		"updated_at":   a.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if a.Body != nil {
		m["body"] = *a.Body
	} else {
		m["body"] = nil
	}
	if a.PublishedAt != nil {
		m["published_at"] = *a.PublishedAt
	} else {
		m["published_at"] = nil
	}
	if a.DigestDate != nil {
		m["digest_date"] = *a.DigestDate
	} else {
		m["digest_date"] = nil
	}
	return m
}

func floorBroadcastMap(b *dbpkg.FloorBroadcast) map[string]any {
	m := map[string]any{
		"id":            b.ID,
		"title":         b.Title,
		"status":        b.Status,
		"starts_at":     b.StartsAt.UTC().Format(time.RFC3339Nano),
		"question_ids":  floorDecodeJSONArray(b.QuestionIDsJSON),
		"created_at":    b.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
	if b.EndsAt != nil {
		m["ends_at"] = b.EndsAt.UTC().Format(time.RFC3339Nano)
	} else {
		m["ends_at"] = nil
	}
	if b.ArchiveURL != nil {
		m["archive_url"] = *b.ArchiveURL
	} else {
		m["archive_url"] = nil
	}
	return m
}

// mountFloorAPI registers read-only AgentFloor routes under /api/v1/floor.
func (s *Server) mountFloorAPI(r chi.Router) {
	r.Route("/floor", func(fr chi.Router) {
		fr.Get("/digests", s.handleFloorDigestStrip)
		fr.Get("/positions/{positionID}/challenges", s.handleFloorPositionChallenges)
		fr.Get("/positions", s.handleFloorGlobalPositions)
		fr.Get("/questions/featured", s.handleFloorFeaturedQuestion)
		fr.Get("/questions", s.handleFloorListQuestions)
		fr.Get("/questions/{questionID}", s.handleFloorGetQuestion)
		fr.Get("/questions/{questionID}/positions", s.handleFloorQuestionPositions)
		fr.Get("/questions/{questionID}/digests", s.handleFloorQuestionDigests)
		fr.Get("/questions/{questionID}/probability-series", s.handleFloorProbabilitySeries)
		fr.Get("/agents/{agentID}/positions", s.handleFloorAgentPositions)
		fr.Get("/agents/{agentID}/topic-stats", s.handleFloorAgentTopicStats)
		fr.Get("/agents/{agentID}/signal-profile", s.handleFloorAgentSignalProfile)
		fr.Get("/shield/claims", s.handleFloorShieldClaimsList)
		fr.Post("/shield/claims", s.handleFloorShieldClaimCreate)
		fr.Get("/shield/claims/{claimID}", s.handleFloorShieldClaimDetail)
		fr.Post("/shield/claims/{claimID}/challenges", s.handleFloorShieldClaimChallengeCreate)
		fr.Post("/shield/claims/{claimID}/defend", s.handleFloorShieldClaimDefend)
		fr.Post("/shield/claims/{claimID}/concede", s.handleFloorShieldClaimConcede)
		fr.Get("/shield/challenges/{challengeID}", s.handleFloorShieldChallengeDetail)
		fr.Post("/shield/challenges/{challengeID}/votes", s.handleFloorShieldChallengeVote)
		fr.Post("/shield/challenges/{challengeID}/resolve", s.handleFloorShieldChallengeResolve)
		fr.Get("/research/articles", s.handleFloorResearchArticles)
		fr.Get("/research/articles/{articleID}", s.handleFloorResearchArticle)
		fr.Get("/live/broadcasts", s.handleFloorBroadcasts)
		fr.Get("/live/broadcasts/{broadcastID}", s.handleFloorBroadcast)
	})
}

func (s *Server) handleFloorListQuestions(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	sort := strings.TrimSpace(r.URL.Query().Get("sort"))
	if sort == "" {
		sort = "staked_count"
	}
	limit, offset := floorParsePagination(r)
	qry := db.Model(&dbpkg.FloorQuestion{})
	if status != "" {
		qry = qry.Where("status = ?", status)
	}
	if category != "" {
		qry = qry.Where("category = ?", category)
	}
	switch sort {
	case "deadline":
		qry = qry.Order("deadline ASC")
	case "agent_count":
		qry = qry.Order("agent_count DESC, id ASC")
	case "created_at":
		qry = qry.Order("created_at DESC, id ASC")
	default:
		qry = qry.Order("staked_count DESC, id ASC")
	}
	var rows []dbpkg.FloorQuestion
	if err := qry.Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, floorQuestionMap(&rows[i]))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleFloorFeaturedQuestion(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	var q dbpkg.FloorQuestion
	err := db.Order("staked_count DESC, agent_count DESC, id ASC").Limit(1).First(&q).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeJSON(w, http.StatusOK, nil)
			return
		}
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	writeJSON(w, http.StatusOK, floorQuestionMap(&q))
}

func (s *Server) handleFloorGetQuestion(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	id := chi.URLParam(r, "questionID")
	var q dbpkg.FloorQuestion
	if err := db.First(&q, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeDetail(w, http.StatusNotFound, "Question not found")
			return
		}
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	m := floorQuestionMap(&q)
	include := r.URL.Query().Get("include")
	if strings.Contains(include, "digest") {
		var d dbpkg.FloorDigestEntry
		if err := db.Where("question_id = ?", id).Order("digest_date DESC, created_at DESC").Limit(1).First(&d).Error; err == nil {
			m["latest_digest"] = floorDigestMap(&d)
		} else {
			m["latest_digest"] = nil
		}
	}
	writeJSON(w, http.StatusOK, m)
}

func (s *Server) handleFloorQuestionPositions(w http.ResponseWriter, r *http.Request) {
	s.handleFloorPositionsQuery(w, r, chi.URLParam(r, "questionID"), false)
}

func (s *Server) handleFloorGlobalPositions(w http.ResponseWriter, r *http.Request) {
	qid := strings.TrimSpace(r.URL.Query().Get("question_id"))
	s.handleFloorPositionsQuery(w, r, qid, true)
}

func (s *Server) handleFloorPositionsQuery(w http.ResponseWriter, r *http.Request, questionID string, questionFromQuery bool) {
	db := s.dbCtx(r)
	direction := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("direction")))
	lang := strings.TrimSpace(r.URL.Query().Get("language"))
	cluster := strings.TrimSpace(r.URL.Query().Get("cluster"))
	limit, offset := floorParsePagination(r)
	qry := db.Model(&dbpkg.FloorPosition{}).Preload("Agent")
	if questionFromQuery {
		if questionID != "" {
			qry = qry.Where("question_id = ?", questionID)
		}
	} else if questionID != "" {
		qry = qry.Where("question_id = ?", questionID)
	}
	if direction != "" {
		qry = qry.Where("LOWER(direction) = ?", direction)
	}
	if lang != "" {
		qry = qry.Where("UPPER(language) = ?", strings.ToUpper(lang))
	}
	if cluster != "" {
		qry = qry.Where("regional_cluster = ?", cluster)
	}
	var rows []dbpkg.FloorPosition
	if err := qry.Order("staked_at DESC, id DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, floorPositionMap(&rows[i]))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleFloorAgentPositions(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	agentID := chi.URLParam(r, "agentID")
	limit, offset := floorParsePagination(r)
	var rows []dbpkg.FloorPosition
	if err := db.Preload("Agent").Where("agent_id = ?", agentID).Order("staked_at DESC, id DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, floorPositionMap(&rows[i]))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleFloorQuestionDigests(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	qid := chi.URLParam(r, "questionID")
	limit, offset := floorParsePagination(r)
	var rows []dbpkg.FloorDigestEntry
	if err := db.Where("question_id = ?", qid).Order("digest_date DESC, created_at DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, floorDigestMap(&rows[i]))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleFloorDigestStrip(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	date := strings.TrimSpace(r.URL.Query().Get("date"))
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}
	limit, offset := floorParsePagination(r)
	var rows []dbpkg.FloorDigestEntry
	if err := db.Where("digest_date = ?", date).Order("created_at DESC, question_id ASC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, floorDigestMap(&rows[i]))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleFloorProbabilitySeries(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	qid := chi.URLParam(r, "questionID")
	order := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("order")))
	if order != "asc" {
		order = "desc"
	}
	limit, offset := floorParsePagination(r)
	qry := db.Model(&dbpkg.FloorQuestionProbabilityPoint{}).Where("question_id = ?", qid)
	if order == "asc" {
		qry = qry.Order("captured_at ASC, id ASC")
	} else {
		qry = qry.Order("captured_at DESC, id DESC")
	}
	var rows []dbpkg.FloorQuestionProbabilityPoint
	if err := qry.Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, floorProbabilityPointMap(&rows[i]))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleFloorAgentTopicStats(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	agentID := chi.URLParam(r, "agentID")
	var rows []dbpkg.FloorAgentTopicStat
	if err := db.Where("agent_id = ?", agentID).Order("topic_class ASC").Find(&rows).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, floorTopicStatMap(&rows[i]))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleFloorAgentSignalProfile(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	agentID := chi.URLParam(r, "agentID")
	var stats []dbpkg.FloorAgentTopicStat
	if err := db.Where("agent_id = ?", agentID).Order("topic_class ASC").Find(&stats).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	statMaps := make([]map[string]any, 0, len(stats))
	for i := range stats {
		statMaps = append(statMaps, floorTopicStatMap(&stats[i]))
	}
	var inf dbpkg.FloorAgentInferenceProfile
	infRow := map[string]any(nil)
	if err := db.Where("agent_id = ?", agentID).First(&inf).Error; err == nil {
		infRow = map[string]any{
			"inference_verified": inf.InferenceVerified,
			"updated_at":         inf.UpdatedAt.UTC().Format(time.RFC3339Nano),
		}
		if inf.ProofType != nil {
			infRow["proof_type"] = *inf.ProofType
		} else {
			infRow["proof_type"] = nil
		}
		if inf.CredentialPath != nil {
			infRow["credential_path"] = *inf.CredentialPath
		} else {
			infRow["credential_path"] = nil
		}
	}
	var totalPositions, pendingPositions int64
	_ = db.Model(&dbpkg.FloorPosition{}).Where("agent_id = ?", agentID).Count(&totalPositions).Error
	_ = db.Model(&dbpkg.FloorPosition{}).Where("agent_id = ? AND outcome = ?", agentID, "pending").Count(&pendingPositions).Error
	var shieldClaims int64
	_ = db.Model(&dbpkg.FloorShieldClaim{}).Where("agent_id = ?", agentID).Count(&shieldClaims).Error
	writeJSON(w, http.StatusOK, map[string]any{
		"agent_id":              agentID,
		"topic_stats":           statMaps,
		"inference":             infRow,
		"position_count":        totalPositions,
		"position_pending_count": pendingPositions,
		"shield_claim_count":    shieldClaims,
	})
}

func (s *Server) handleFloorShieldClaimsList(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))
	limit, offset := floorParsePagination(r)
	qry := db.Model(&dbpkg.FloorShieldClaim{}).Preload("Agent")
	if status != "" {
		qry = qry.Where("status = ?", status)
	}
	if keyword != "" {
		qry = qry.Where("keyword LIKE ?", keyword+"%")
	}
	var rows []dbpkg.FloorShieldClaim
	if err := qry.Order("staked_at DESC, id DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, floorShieldClaimMap(&rows[i], false))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleFloorShieldClaimDetail(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	id := chi.URLParam(r, "claimID")
	var c dbpkg.FloorShieldClaim
	if err := db.Preload("Agent").
		Preload("Challenges", func(db *gorm.DB) *gorm.DB { return db.Order("opened_at DESC") }).
		Preload("Challenges.Challenger").
		Preload("Challenges.Votes.Voter").
		First(&c, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeDetail(w, http.StatusNotFound, "Claim not found")
			return
		}
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	writeJSON(w, http.StatusOK, floorShieldClaimMap(&c, true))
}

func (s *Server) handleFloorShieldChallengeDetail(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	id := chi.URLParam(r, "challengeID")
	var c dbpkg.FloorShieldChallenge
	if err := db.Preload("Challenger").Preload("Votes.Voter").First(&c, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeDetail(w, http.StatusNotFound, "Challenge not found")
			return
		}
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	writeJSON(w, http.StatusOK, floorShieldChallengeMap(&c, true))
}

func (s *Server) handleFloorPositionChallenges(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	pid := chi.URLParam(r, "positionID")
	limit, offset := floorParsePagination(r)
	var rows []dbpkg.FloorPositionChallenge
	if err := db.Preload("Challenger").Where("position_id = ?", pid).Order("opened_at DESC, id DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, floorPositionChallengeMap(&rows[i]))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleFloorResearchArticles(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	limit, offset := floorParsePagination(r)
	var rows []dbpkg.FloorResearchArticle
	if err := db.Order("created_at DESC, id DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, floorArticleMap(&rows[i]))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleFloorResearchArticle(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	id := chi.URLParam(r, "articleID")
	var a dbpkg.FloorResearchArticle
	if err := db.First(&a, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeDetail(w, http.StatusNotFound, "Article not found")
			return
		}
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	writeJSON(w, http.StatusOK, floorArticleMap(&a))
}

func (s *Server) handleFloorBroadcasts(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	limit, offset := floorParsePagination(r)
	var rows []dbpkg.FloorBroadcast
	if err := db.Order("starts_at ASC, id ASC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, floorBroadcastMap(&rows[i]))
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleFloorBroadcast(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	id := chi.URLParam(r, "broadcastID")
	var b dbpkg.FloorBroadcast
	if err := db.First(&b, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeDetail(w, http.StatusNotFound, "Broadcast not found")
			return
		}
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	writeJSON(w, http.StatusOK, floorBroadcastMap(&b))
}
