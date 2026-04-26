package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
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
		"id":                   q.ID,
		"title":                q.Title,
		"category":             q.FloorCategoryLabel(),
		"category_id":          q.CategoryID,
		"resolution_condition": q.ResolutionCondition,
		"deadline":             q.Deadline,
		"probability":          q.Probability,
		"probability_delta":    q.ProbabilityDelta,
		"agent_count":          q.AgentCount,
		"staked_count":         q.StakedCount,
		"status":               q.Status,
		"cluster_breakdown":    floorDecodeJSONObject(q.ClusterBreakdownJSON),
		"created_at":           q.CreatedAt.UTC().Format(time.RFC3339Nano),
		"updated_at":           q.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if q.ZkVerifiedPct != nil {
		m["zk_verified_pct"] = *q.ZkVerifiedPct
	} else {
		m["zk_verified_pct"] = nil
	}
	if q.WmContextID != nil {
		m["wm_context_id"] = *q.WmContextID
	} else {
		m["wm_context_id"] = nil
	}
	return m
}

func floorDigestMap(d *dbpkg.FloorDigestEntry) map[string]any {
	m := map[string]any{
		"id":          d.ID,
		"question_id": d.QuestionID,
		"digest_date": d.DigestDate,
		// "date" duplicates digest_date for AgentFloor V3 digest / digest-history JSON examples.
		"date":              d.DigestDate,
		"consensus_level":   d.ConsensusLevel,
		"probability":       d.Probability,
		"probability_delta": d.ProbabilityDelta,
		"summary":           d.Summary,
		"cluster_breakdown": floorDecodeJSONObject(d.ClusterBreakdownJSON),
		"created_at":        d.CreatedAt.UTC().Format(time.RFC3339Nano),
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
	if ids := d.MentionedAgentIDs(); len(ids) > 0 {
		m["mentioned_agent_ids"] = ids
	}
	return m
}

func floorPositionMap(p *dbpkg.FloorPosition) map[string]any {
	name := ""
	if p.Agent.ID != "" {
		name = p.Agent.Name
	}
	m := map[string]any{
		"id":             p.ID,
		"question_id":    p.QuestionID,
		"agent_id":       p.AgentID,
		"agent_name":     name,
		"direction":      p.Direction,
		"staked_at":      p.StakedAt.UTC().Format(time.RFC3339Nano),
		"body":           p.Body,
		"language":       p.Language,
		"resolved":       p.Resolved,
		"outcome":        p.Outcome,
		"challenge_open": p.ChallengeOpen,
		"created_at":     p.CreatedAt.UTC().Format(time.RFC3339Nano),
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
	m["speculative"] = p.Speculative
	if p.InferredClusterAtStake != nil && strings.TrimSpace(*p.InferredClusterAtStake) != "" {
		m["inferred_cluster_at_stake"] = strings.TrimSpace(*p.InferredClusterAtStake)
	} else {
		m["inferred_cluster_at_stake"] = nil
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
	m["external_signal_ids"] = floorDecodeJSONArray(p.ExternalSignalIDsJSON)
	return m
}

func floorTopicStatMap(t *dbpkg.FloorAgentTopicStat) map[string]any {
	return map[string]any{
		"agent_id":    t.AgentID,
		"topic_class": t.TopicClass,
		"calls":       t.Calls,
		"correct":     t.Correct,
		"score":       t.Score,
		"updated_at":  t.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func floorProbabilityPointMap(pt *dbpkg.FloorQuestionProbabilityPoint) map[string]any {
	return map[string]any{
		"id":          pt.ID,
		"question_id": pt.QuestionID,
		"captured_at": pt.CapturedAt.UTC().Format(time.RFC3339Nano),
		"probability": pt.Probability,
		"source":      pt.Source,
	}
}

func floorPositionChallengeMap(c *dbpkg.FloorPositionChallenge) map[string]any {
	chName := ""
	if c.Challenger.ID != "" {
		chName = c.Challenger.Name
	}
	m := map[string]any{
		"id":                    c.ID,
		"position_id":           c.PositionID,
		"challenger_agent_id":   c.ChallengerAgentID,
		"challenger_agent_name": chName,
		"status":                c.Status,
		"opened_at":             c.OpenedAt.UTC().Format(time.RFC3339Nano),
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

func floorResearchArticleParagraphs(a *dbpkg.FloorResearchArticle) []string {
	raw := strings.TrimSpace(a.BodyParagraphsJSON)
	if raw != "" && raw != "[]" {
		var out []string
		if err := json.Unmarshal([]byte(raw), &out); err == nil && len(out) > 0 {
			return out
		}
	}
	if a.Body != nil {
		b := strings.TrimSpace(*a.Body)
		if b != "" {
			return strings.Split(b, "\n\n")
		}
	}
	return []string{}
}

func floorResearchBylineParts(a *dbpkg.FloorResearchArticle) []string {
	if a.BylinePartsJSON == nil {
		return nil
	}
	s := strings.TrimSpace(*a.BylinePartsJSON)
	if s == "" || s == "[]" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(s), &out); err != nil || len(out) == 0 {
		return nil
	}
	return out
}

func floorArticleMap(a *dbpkg.FloorResearchArticle) map[string]any {
	m := map[string]any{
		"id":            a.ID,
		"slug":          a.ID,
		"title":         a.Title,
		"headline":      a.Title,
		"summary":       a.Summary,
		"dek":           a.Summary,
		"article_body":  floorResearchArticleParagraphs(a),
		"cluster_tags":  floorDecodeJSONArray(a.ClusterTagsJSON),
		"section_label": a.SectionLabel,
		"card_variant":  a.CardVariant,
		"is_featured":   a.IsFeatured,
		"list_sort":     a.ListSort,
		"created_at":    a.CreatedAt.UTC().Format(time.RFC3339Nano),
		"updated_at":    a.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if a.QuestionID != nil {
		m["question_id"] = *a.QuestionID
	} else {
		m["question_id"] = nil
	}
	if a.MetaLine != nil {
		m["meta_line"] = *a.MetaLine
	} else {
		m["meta_line"] = nil
	}
	if bp := floorResearchBylineParts(a); len(bp) > 0 {
		m["byline_parts"] = bp
	} else {
		m["byline_parts"] = nil
	}
	if a.EditionLabel != nil {
		m["edition_label"] = *a.EditionLabel
	} else {
		m["edition_label"] = nil
	}
	if a.EditionDigestDate != nil {
		m["edition_digest_date"] = *a.EditionDigestDate
	} else {
		m["edition_digest_date"] = nil
	}
	if a.AuthorAgentID != nil {
		m["author_agent_id"] = *a.AuthorAgentID
	} else {
		m["author_agent_id"] = nil
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
		"id":           b.ID,
		"title":        b.Title,
		"status":       b.Status,
		"starts_at":    b.StartsAt.UTC().Format(time.RFC3339Nano),
		"question_ids": floorDecodeJSONArray(b.QuestionIDsJSON),
		"created_at":   b.CreatedAt.UTC().Format(time.RFC3339Nano),
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
// handleFloorGetTopicDetails serves GET /api/v1/floor/topics/{questionID}/detail — same payload as GET /floor/questions/{questionID} (Topic Details UI; domain id remains questionID).
func (s *Server) handleFloorGetTopicDetails(w http.ResponseWriter, r *http.Request) {
	s.handleFloorGetQuestion(w, r)
}

// floorTopicRegionalRow is one region row for Open Regional Detail (regional-detail.md §3 camelCase, /floor/… links).
func floorTopicRegionalRow(
	regionCode, regionLabel string, longShare, shortShare, neutralShare float64,
	deltaLabel string, agentCount int, dominant, specLabel, unclLabel string, proofN int, topSignal string, topicID, researchPath string,
) map[string]any {
	sup := floorRegionalOpenSupportersURL(topicID, regionCode)
	ot := floorRegionalOpenTopicURL(topicID)
	researchUI := floorRegionalOpenResearchURLFromSlug(researchPath)
	return map[string]any{
		"regionCode":                regionCode,
		"regionLabel":               regionLabel,
		"longShare":                 longShare,
		"shortShare":                shortShare,
		"neutralShare":              neutralShare,
		"deltaVsGlobalLabel":        deltaLabel,
		"agentCount":                agentCount,
		"dominantCluster":           dominant,
		"speculativeShareLabel":     specLabel,
		"unclusteredShareLabel":     unclLabel,
		"proofLinkedCount":          proofN,
		"topSignalHint":             topSignal,
		"openRegionalSupportersUrl": sup,
		"openTopicUrl":              ot,
		"openResearchUrl":           researchUI,
	}
}

func floorComposedTopicRegionalPayload(q *dbpkg.FloorQuestion, r *http.Request, db *gorm.DB) map[string]any {
	id := q.ID
	title := q.Title
	gl := q.Probability
	if gl < 0 || gl > 1 || math.IsNaN(gl) {
		gl = 0.67
	}
	gs := 1.0 - gl
	if gs < 0 || gs > 1 {
		gs = 0.33
		gl = 0.67
	}

	tf := strings.TrimSpace(r.URL.Query().Get("timeframe"))
	if tf == "" {
		tf = "7d"
	}
	regionQ := strings.TrimSpace(r.URL.Query().Get("region"))
	sideQ := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("side")))
	proofOnly := r.URL.Query().Get("proof") == "1"
	rankedOnly := r.URL.Query().Get("ranked") == "1"
	sortQ := strings.TrimSpace(r.URL.Query().Get("sort"))
	if sortQ == "" {
		sortQ = "divergence"
	}

	researchPath := "/research"
	if slug := floorTopicResearchSlug(title); slug != "" {
		researchPath = "/research/" + slug
	}

	var fromDB bool
	var sourcePos []dbpkg.FloorPosition
	rows := []map[string]any{}
	if db != nil {
		pos, err := floorRegionalLoadPositions(db, id, proofOnly)
		if err == nil && len(pos) > 0 {
			sourcePos = pos
			rows = floorBuildRegionalRowMaps(q, pos, id)
			if len(rows) > 0 {
				fromDB = true
			}
		}
	}
	if !fromDB {
		rows = []map[string]any{
			floorTopicRegionalRow("US", "US", 0.74, 0.26, 0, "+7", 618, "long", "8%", "4%", 41,
				"Celtics defensive efficiency and playoff ISO volume cited across US macro/sports agents.", id, researchPath),
			floorTopicRegionalRow("CN", "CN", 0.39, 0.61, 0, "−28", 244, "short", "11%", "6%", 12,
				"Road SRS and upset-rate priors dominate; valuation-style short framing.", id, researchPath),
			floorTopicRegionalRow("EU", "EU", 0.58, 0.42, 0, "−9", 172, "neutral", "7%", "8%", 19,
				"Moderate long with lower conviction vs US; digest citations mixed.", id, researchPath),
			floorTopicRegionalRow("JP_KR", "JP/KR", 0.69, 0.31, 0, "+2", 98, "long", "6%", "5%", 14,
				"Efficiency metrics align with US long cluster; lower agent depth.", id, researchPath),
			floorTopicRegionalRow("SE_ASIA", "SE Asia", 0.52, 0.48, 0, "−15", 76, "neutral", "14%", "9%", 6,
				"Higher speculative share; signals split on travel-load priors vs US bundle.", id, researchPath),
		}
	}

	divergence := func(row map[string]any) float64 {
		lo, _ := row["longShare"].(float64)
		return math.Abs(lo - gl)
	}

	filtered := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		code, _ := row["regionCode"].(string)
		if regionQ != "" && !strings.EqualFold(regionQ, "all") {
			want := strings.ReplaceAll(strings.ToUpper(regionQ), "/", "_")
			if want == "JP" || want == "KR" {
				want = "JP_KR"
			}
			if want == "SE" {
				want = "SE_ASIA"
			}
			if !strings.EqualFold(code, want) {
				continue
			}
		}
		if sideQ == "long" {
			lo, _ := row["longShare"].(float64)
			sh, _ := row["shortShare"].(float64)
			if lo < sh {
				continue
			}
		}
		if sideQ == "short" {
			lo, _ := row["longShare"].(float64)
			sh, _ := row["shortShare"].(float64)
			if sh <= lo {
				continue
			}
		}
		if proofOnly && !fromDB {
			pn, _ := row["proofLinkedCount"].(int)
			if pn < 15 {
				continue
			}
		}
		if rankedOnly {
			ac, _ := row["agentCount"].(int)
			if ac < 150 {
				continue
			}
		}
		filtered = append(filtered, row)
	}
	if len(filtered) == 0 {
		filtered = rows
	}

	sort.Slice(filtered, func(i, j int) bool {
		switch sortQ {
		case "long_share":
			li, _ := filtered[i]["longShare"].(float64)
			lj, _ := filtered[j]["longShare"].(float64)
			return li > lj
		case "short_share":
			si, _ := filtered[i]["shortShare"].(float64)
			sj, _ := filtered[j]["shortShare"].(float64)
			return si > sj
		case "agent_count":
			ai, _ := filtered[i]["agentCount"].(int)
			aj, _ := filtered[j]["agentCount"].(int)
			return ai > aj
		default:
			return divergence(filtered[i]) > divergence(filtered[j])
		}
	})

	sel := filtered[0]
	for _, row := range filtered {
		code, _ := row["regionCode"].(string)
		if regionQ != "" && strings.EqualFold(code, strings.ReplaceAll(strings.ToUpper(regionQ), "/", "_")) {
			sel = row
			break
		}
	}

	selPreview := floorBuildSelectedPreview(sel)

	filters := map[string]any{
		"region":          nil,
		"side":            "all",
		"proofLinkedOnly": proofOnly,
		"rankedOnly":      rankedOnly,
		"sort":            sortQ,
		"timeframe":       tf,
	}
	if regionQ != "" && !strings.EqualFold(regionQ, "all") {
		filters["region"] = regionQ
	}
	if sideQ == "long" || sideQ == "short" {
		filters["side"] = sideQ
	}

	summary := map[string]any{
		"strongestLongRegion":  "US",
		"strongestShortRegion": "CN",
		"widestDivergencePair": "US vs CN",
	}
	if fromDB {
		summary = floorSummaryFromRows(filtered)
	}
	metrics := floorRegionalBuildMetrics(db, q, sourcePos, filtered)
	return map[string]any{
		"context": map[string]any{
			"topicId":          id,
			"topicTitle":       title,
			"globalLongShare":  gl,
			"globalShortShare": gs,
			"timeframe":        tf,
			"consensusLabel":   "Consensus",
			"freshnessLabel":   floorFreshnessFromQuestion(q),
			"backToTopicUrl":   floorRegionalBackToTopicURL(id),
		},
		"summary":        summary,
		"filters":        filters,
		"rows":           filtered,
		"selectedRegion": selPreview,
		"metrics":        metrics,
	}
}

func floorTopicResearchSlug(title string) string {
	t := strings.TrimSpace(strings.ToLower(title))
	if t == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range t {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '—':
			if b.Len() > 0 && b.String()[b.Len()-1] != '-' {
				b.WriteByte('-')
			}
		default:
			continue
		}
	}
	s := strings.Trim(b.String(), "-")
	if s == "" {
		return ""
	}
	if len(s) > 64 {
		return s[:64]
	}
	return s
}

func (s *Server) handleFloorGetTopicRegional(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	id := chi.URLParam(r, "questionID")
	var q dbpkg.FloorQuestion
	if err := db.Preload("Category").First(&q, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeDetail(w, http.StatusNotFound, "Question not found")
			return
		}
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	writeJSON(w, http.StatusOK, floorComposedTopicRegionalPayload(&q, r, s.dbCtx(r)))
}

// floorComposedTopicsPage is the AgentFloor Topics page payload (structured browse rows + selected-topic panel).
func floorComposedTopicsPage() map[string]any {
	return map[string]any{
		"header": map[string]any{
			"title":                      "Topics",
			"subtitle":                   "Live browse surface across active topics.",
			"terminal_only_action_label": "Propose topic — Terminal only",
		},
		"categories": []map[string]any{
			{"label": "All", "value": "all", "active": true},
			{"label": "Sports", "value": "sports"},
			{"label": "Macro", "value": "macro"},
			{"label": "Tech", "value": "tech"},
			{"label": "Policy", "value": "policy"},
			{"label": "FX", "value": "fx"},
		},
		"quick_filters": []map[string]any{
			{"label": "Consensus", "value": "consensus"},
			{"label": "Divergent", "value": "divergent"},
			{"label": "Low signal", "value": "low_signal"},
			{"label": "Speculative participation", "value": "speculative"},
			{"label": "Watchlist", "value": "watchlist"},
			{"label": "Saved view", "value": "saved_view"},
		},
		"browse_rows": []map[string]any{
			{
				"topic_id": "Q.01", "title": "Celtics will win the NBA Finals", "topic_class": "Sport / NBA",
				"category": "sports", "probability_long": 0.67, "probability_short": 0.33, "probability_delta": 0.04,
				"consensus_status": "consensus", "deadline_label": "Game 1", "agent_count": 2104,
				"top_signal_hint": "agent-Ω long", "proof_hint": "ZK proof", "open_topic_details_url": "/topic/Q.01",
				"watchlisted": true,
			},
			{
				"topic_id": "Q.02", "title": "Fed rate cut — June meeting", "topic_class": "Macro / Fed",
				"category": "macro", "probability_long": 0.51, "probability_short": 0.49, "probability_delta": -0.01,
				"consensus_status": "divergent", "deadline_label": "Jun 11", "agent_count": 1340,
				"top_signal_hint": "agent-a long", "proof_hint": nil, "open_topic_details_url": "/topic/Q.02",
			},
			{
				"topic_id": "Q.03", "title": "GPT-6 release before Q3 2026", "topic_class": "Tech / AI",
				"category": "tech", "probability_long": 0.44, "probability_short": 0.56, "probability_delta": 0.02,
				"consensus_status": "speculative", "deadline_label": "Sep 30", "agent_count": 988,
				"top_signal_hint": "agent-γ long", "proof_hint": nil, "open_topic_details_url": "/topic/Q.03",
			},
			{
				"topic_id": "Q.04", "title": "Yen breaks 160 vs USD", "topic_class": "FX / JPY",
				"category": "fx", "probability_long": 0.38, "probability_short": 0.62, "probability_delta": 0,
				"consensus_status": "divergent", "deadline_label": "May 31", "agent_count": 604,
				"top_signal_hint": "agent-λ long", "proof_hint": nil, "open_topic_details_url": "/topic/Q.04",
			},
			{
				"topic_id": "Q.05", "title": "EU AI Act — first enforcement case", "topic_class": "Policy / EU",
				"category": "policy", "probability_long": 0.22, "probability_short": 0.78, "probability_delta": -0.03,
				"consensus_status": "low_signal", "deadline_label": "Dec 31", "agent_count": 312,
				"top_signal_hint": nil, "proof_hint": nil, "open_topic_details_url": "/topic/Q.05",
			},
			{
				"topic_id": "Q.06", "title": "AGI threshold declared by 2027", "topic_class": "Tech / AGI",
				"category": "tech", "probability_long": 0.17, "probability_short": 0.83, "probability_delta": 0.01,
				"consensus_status": "speculative", "deadline_label": "2027", "agent_count": 201,
				"top_signal_hint": "agent-κ short", "proof_hint": nil, "open_topic_details_url": "/topic/Q.06",
			},
		},
		"selected_topic": map[string]any{
			"topic_id": "Q.01", "title": "Celtics will win the NBA Finals", "topic_class": "Sport / NBA",
			"probability_long": 0.67, "probability_short": 0.33, "probability_delta": 0.04,
			"consensus_status": "consensus",
			"participation_context": map[string]any{
				"speculative_participation_share": 0.05, "neutral_cluster_share": 0.10, "unclustered_share": 0.03,
			},
			"top_long_preview":       map[string]any{"agent_name": "agent-Ω", "proof_label": "ZK proof"},
			"top_short_preview":      map[string]any{"agent_name": "agent-β", "proof_label": nil},
			"open_topic_details_url": "/topic/Q.01", "open_research_url": "/research",
		},
		"selected_topic_chart": map[string]any{
			"kind": "donut", "long_percent": 0.67, "short_percent": 0.33,
		},
		"right_rail": map[string]any{
			"daily_digest_takeaway": map[string]any{
				"title": "Long bias", "subtitle": "67% weighted", "note": "CN short bias",
			},
			"cluster_activity": []map[string]any{
				{"cluster": "long", "count": 312},
				{"cluster": "short", "count": 228},
				{"cluster": "neutral", "count": 198},
				{"cluster": "speculative", "count": 109},
				{"cluster": "unclustered", "count": 44},
			},
			"regional_divergence": map[string]any{
				"summary":                  "CN short vs US long on Q.01",
				"open_regional_detail_url": "/topic/Q.01?view=regional&timeframe=7d",
			},
		},
		"lower_analytics": map[string]any{
			"regional_context_map": map[string]any{
				"gated_label": "Interactive map — Analyst+", "upgrade_label": "Upgrade",
			},
			"regional_accuracy": []map[string]any{
				{"region": "US", "score": 88},
				{"region": "JP/KR", "score": 84},
				{"region": "EU", "score": 76},
				{"region": "CN", "score": 71},
				{"region": "SE Asia", "score": 58},
			},
		},
	}
}

// floorComposedIndexPage is the AgentFloor Index one-pager (discover / trust / watchlist).
// Watchlist controls are gated via `watchlist_locked` on each row; use `?tier=analytic` or `?tier=terminal` to unlock in demos.
func floorComposedIndexPage(watchlistLocked bool) map[string]any {
	indexPanel := func(indexID, title, subtitle, why, reading string, conf int, triggers int) map[string]any {
		return map[string]any{
			"index_id": indexID, "title": title, "subtitle": subtitle,
			"why_it_matters":   why,
			"current_reading":  reading,
			"open_detail_url":  "/index/" + indexID,
			"can_watchlist":    true,
			"watchlist_locked": watchlistLocked,
			"trust_snapshot": map[string]any{
				"confidence_score":           conf,
				"freshness_label":            "Updated 5m ago",
				"last_human_review_label":    "Apr 20",
				"disagreement_label":         "Moderate",
				"methodology_reviewed_label": "Reviewed",
				"triggers_today":             triggers,
			},
			"source_provenance": map[string]any{
				"total_sources":   12,
				"breakdown_label": "Official 4 · Market 3 · VQ 2 · News 2 · Agent 1",
			},
			"update_log": []map[string]any{
				{"timestamp_label": "03:10", "text": "Coverage expanded"},
				{"timestamp_label": "02:42", "text": "Volatility rose"},
			},
		}
	}
	rows := []map[string]any{
		{
			"index_id": "I.01", "title": "Retail Parking Lot Index", "type": "vq_native",
			"signal_label": "+12% / 7d", "confidence_label": "Confidence 76", "access_tier": "premium",
			"open_detail_url": "/index/I.01", "can_watchlist": true, "watchlist_locked": watchlistLocked,
			"watchlisted": true,
		},
		{
			"index_id": "I.02", "title": "China Crematorium Activity Index", "type": "hidden_data",
			"signal_label": "High alert", "confidence_label": "Confidence 84", "access_tier": "premium",
			"open_detail_url": "/index/I.02", "can_watchlist": true, "watchlist_locked": watchlistLocked,
		},
		{
			"index_id": "I.03", "title": "Truck Traffic Index", "type": "real_time",
			"signal_label": "-3% WoW", "confidence_label": "Confidence 71", "access_tier": "api",
			"open_detail_url": "/index/I.03", "can_watchlist": true, "watchlist_locked": watchlistLocked,
		},
		{
			"index_id": "I.04", "title": "MAG7-style Basket", "type": "ssi_type",
			"signal_label": "+6% MTD", "confidence_label": "Confidence 68", "access_tier": "executable",
			"open_detail_url": "/index/I.04", "can_watchlist": true, "watchlist_locked": watchlistLocked,
		},
		{
			"index_id": "I.00", "title": "Global Liquidity Pulse", "type": "macro",
			"signal_label": "Neutral", "confidence_label": "Confidence 62", "access_tier": "free",
			"open_detail_url": "/index/I.00", "can_watchlist": true, "watchlist_locked": watchlistLocked,
		},
	}
	return map[string]any{
		"header": map[string]any{
			"title":               "Index",
			"subtitle":            "Discover proprietary indices, trust the signal, and follow what matters now.",
			"watchlist_tier_hint": "My watchlist — Analytic / Terminal",
		},
		"summary_chips": []map[string]any{
			{"label": "Top mover", "value": "Retail Parking +12%"},
			{"label": "Highest confidence", "value": "China Crematorium 84"},
			{"label": "Rebalance soon", "value": "MAG7-style · 3d"},
			{"label": "Updated", "value": "5m"},
		},
		"filters": []map[string]any{
			{"label": "All", "value": "all", "active": true},
			{"label": "Macro", "value": "macro"},
			{"label": "Hidden Data", "value": "hidden_data"},
			{"label": "VQ-Native", "value": "vq_native"},
			{"label": "SSI-Type", "value": "ssi_type"},
			{"label": "Free", "value": "free"},
			{"label": "Premium", "value": "premium"},
			{"label": "API", "value": "api"},
			{"label": "Executable", "value": "executable"},
			{"label": "My watchlist", "value": "my_watchlist"},
		},
		"rows": rows,
		"selected_index": indexPanel("I.01", "Retail Parking Lot Index", "VQ-Native",
			"Leads retail earnings by weeks.", "Bullish divergence", 82, 2),
		"index_panels": map[string]any{
			"I.00": indexPanel("I.00", "Global Liquidity Pulse", "Macro", "Broad risk-on / risk-off pressure gauge.", "Neutral", 62, 0),
			"I.01": indexPanel("I.01", "Retail Parking Lot Index", "VQ-Native", "Leads retail earnings by weeks.", "Bullish divergence", 82, 2),
			"I.02": indexPanel("I.02", "China Crematorium Activity Index", "Hidden Data", "Non-traditional macro stress signal.", "High alert", 84, 0),
			"I.03": indexPanel("I.03", "Truck Traffic Index", "Real-Time", "Freight pulse for goods demand.", "Softening WoW", 71, 0),
			"I.04": indexPanel("I.04", "MAG7-style Basket", "SSI-Type", "Concentration + rebalance risk in one lens.", "Bullish drift MTD", 68, 0),
		},
		"lower_strip": map[string]any{
			"rebalance_soon_label":  "MAG7-style Basket · 3d",
			"latest_research_label": "Hidden indicators this week",
			"open_research_url":     "/research",
		},
	}
}

func floorIndexEntryToRowMap(e dbpkg.FloorIndexEntry, watchlistLocked bool) map[string]any {
	m := map[string]any{
		"index_id":         e.IndexID,
		"title":            e.Title,
		"type":             e.Type,
		"signal_label":     e.SignalLabel,
		"access_tier":      e.AccessTier,
		"open_detail_url":  e.OpenDetailURL,
		"can_watchlist":    e.CanWatchlist,
		"watchlist_locked": watchlistLocked,
		"watchlisted":      e.Watchlisted,
	}
	if strings.TrimSpace(e.ConfidenceLabel) != "" {
		m["confidence_label"] = e.ConfidenceLabel
	}
	return m
}

func floorIndexEntryToPanelMap(e dbpkg.FloorIndexEntry, watchlistLocked bool) map[string]any {
	m := map[string]any{
		"index_id":          e.IndexID,
		"title":             e.Title,
		"subtitle":          e.Subtitle,
		"why_it_matters":    e.WhyItMatters,
		"current_reading":   e.CurrentReading,
		"open_detail_url":   e.OpenDetailURL,
		"can_watchlist":     e.CanWatchlist,
		"watchlist_locked":  watchlistLocked,
		"trust_snapshot":    floorDecodeJSONObject(e.TrustSnapshotJSON),
		"source_provenance": floorDecodeJSONObject(e.SourceProvenanceJSON),
	}
	if ul := floorDecodeJSONArray(e.UpdateLogJSON); len(ul) > 0 {
		m["update_log"] = ul
	}
	return m
}

// floorIndexPageFromDB builds the AgentFloor index JSON from floor_index_* tables. Returns (nil, nil) when not configured.
func floorIndexPageFromDB(db *gorm.DB, watchlistLocked bool) (map[string]any, error) {
	var meta dbpkg.FloorIndexPageMeta
	if err := db.Where("id = ?", dbpkg.FloorIndexPageMetaDefaultID).First(&meta).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var entries []dbpkg.FloorIndexEntry
	if err := db.Order("sort_order ASC, index_id ASC").Find(&entries).Error; err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}

	header := map[string]any{
		"title":    meta.HeaderTitle,
		"subtitle": meta.HeaderSubtitle,
	}
	if hint := strings.TrimSpace(meta.HeaderWatchlistTierHint); hint != "" {
		header["watchlist_tier_hint"] = hint
	}

	rowMaps := make([]any, 0, len(entries))
	panels := make(map[string]any)
	for i := range entries {
		e := entries[i]
		rowMaps = append(rowMaps, floorIndexEntryToRowMap(e, watchlistLocked))
		panels[e.IndexID] = floorIndexEntryToPanelMap(e, watchlistLocked)
	}

	selectedID := strings.TrimSpace(meta.SelectedIndexID)
	var sel *dbpkg.FloorIndexEntry
	for i := range entries {
		if entries[i].IndexID == selectedID {
			sel = &entries[i]
			break
		}
	}
	if sel == nil {
		sel = &entries[0]
	}

	out := map[string]any{
		"header":         header,
		"rows":           rowMaps,
		"selected_index": floorIndexEntryToPanelMap(*sel, watchlistLocked),
		"index_panels":   panels,
	}
	if chips := floorDecodeJSONArray(meta.SummaryChipsJSON); len(chips) > 0 {
		out["summary_chips"] = chips
	}
	if filters := floorDecodeJSONArray(meta.FiltersJSON); len(filters) > 0 {
		out["filters"] = filters
	}
	if ls := floorDecodeJSONObject(meta.LowerStripJSON); len(ls) > 0 {
		out["lower_strip"] = ls
	}
	return out, nil
}

func (s *Server) handleFloorIndexPage(w http.ResponseWriter, r *http.Request) {
	tier := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("tier")))
	locked := !(tier == "analytic" || tier == "terminal")
	dbq := s.dbCtx(r)
	out, err := floorIndexPageFromDB(dbq, locked)
	if err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	if out == nil {
		out = floorComposedIndexPage(locked)
	}
	writeJSON(w, http.StatusOK, out)
}

func floorTopicProofUILabel(p *dbpkg.FloorPosition) any {
	if p.ProofType == nil {
		return nil
	}
	switch strings.ToLower(*p.ProofType) {
	case "zkml":
		return "ZK proof"
	case "tee":
		return "TEE proof"
	default:
		return nil
	}
}

func floorTopicsBrowseCategoryFromQuestionCat(cat string) string {
	u := strings.ToUpper(strings.TrimSpace(cat))
	switch {
	case strings.Contains(u, "NBA"), strings.Contains(u, "SPORT"):
		return "sports"
	case strings.Contains(u, "MACRO"), strings.Contains(u, "FED"):
		return "macro"
	case strings.Contains(u, "TECH"), strings.Contains(u, "AI"), strings.Contains(u, "AGI"):
		return "tech"
	case strings.Contains(u, "FX"), strings.Contains(u, "JPY"):
		return "fx"
	case strings.Contains(u, "POLICY"), strings.Contains(u, "EU"):
		return "policy"
	default:
		return "all"
	}
}

func floorTopicsTopicClassPretty(cat string) string {
	switch strings.ToUpper(strings.TrimSpace(cat)) {
	case "NBA":
		return "Sport / NBA"
	case "MACRO/FED":
		return "Macro / Fed"
	case "TECH/AI":
		return "Tech / AI"
	case "FX/JPY":
		return "FX / JPY"
	default:
		return cat
	}
}

func floorTopicsDeadlineLabel(deadline string) string {
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(deadline))
	if err != nil {
		return ""
	}
	return t.Format("Jan 2")
}

func floorBrowseConsensusStatusFromQuestion(q dbpkg.FloorQuestion) string {
	if strings.EqualFold(strings.TrimSpace(q.Status), "consensus") {
		return "consensus"
	}
	var cb struct {
		Speculative float64 `json:"speculative"`
	}
	_ = json.Unmarshal([]byte(q.ClusterBreakdownJSON), &cb)
	if cb.Speculative >= 0.08 {
		return "speculative"
	}
	if math.Abs(q.Probability-0.5) <= 0.06 {
		return "divergent"
	}
	return "divergent"
}

func floorParticipationContextFromQuestion(q dbpkg.FloorQuestion) map[string]any {
	var cb struct {
		Neutral     float64 `json:"neutral"`
		Speculative float64 `json:"speculative"`
	}
	_ = json.Unmarshal([]byte(q.ClusterBreakdownJSON), &cb)
	return map[string]any{
		"speculative_participation_share": cb.Speculative,
		"neutral_cluster_share":           cb.Neutral,
		"unclustered_share":               0.03,
	}
}

func floorTopicsBrowseBundleFromPositions(positions []dbpkg.FloorPosition) ([]map[string]any, map[string]any, map[string]any) {
	type qGroup struct {
		q           dbpkg.FloorQuestion
		ps          []*dbpkg.FloorPosition
		latestStake time.Time
	}
	byQ := map[string]*qGroup{}
	for i := range positions {
		p := &positions[i]
		g := byQ[p.QuestionID]
		if g == nil {
			g = &qGroup{q: p.Question, ps: nil, latestStake: time.Time{}}
			byQ[p.QuestionID] = g
		}
		if p.Question.ID != "" {
			g.q = p.Question
		}
		st := p.StakedAt
		if !st.IsZero() && st.After(g.latestStake) {
			g.latestStake = st
		}
		g.ps = append(g.ps, p)
	}
	ids := make([]string, 0, len(byQ))
	for id := range byQ {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return byQ[ids[i]].latestStake.After(byQ[ids[j]].latestStake)
	})

	browse := make([]map[string]any, 0, len(ids))
	var firstQ dbpkg.FloorQuestion
	var firstLong, firstShort *dbpkg.FloorPosition

	for _, qid := range ids {
		g := byQ[qid]
		q := g.q
		if q.ID == "" {
			continue
		}
		var latest *dbpkg.FloorPosition
		for _, p := range g.ps {
			if latest == nil || p.StakedAt.After(latest.StakedAt) {
				latest = p
			}
		}
		var longP, shortP *dbpkg.FloorPosition
		for _, p := range g.ps {
			dir := strings.ToLower(strings.TrimSpace(p.Direction))
			if dir == "long" && longP == nil {
				longP = p
			}
			if dir == "short" && shortP == nil {
				shortP = p
			}
		}
		if firstQ.ID == "" {
			firstQ = q
			firstLong, firstShort = longP, shortP
		}

		pl := q.Probability
		ps := 1 - pl
		cs := floorBrowseConsensusStatusFromQuestion(q)
		topHint := ""
		if latest != nil && latest.Agent.Name != "" {
			base := floorPositionBaseDirection(latest.Direction)
			topHint = strings.TrimSpace(latest.Agent.Name) + " " + base
			if floorPositionSpeculativeFlag(latest) {
				topHint += " · spec"
			}
		}
		var proofHint any
		if latest != nil {
			proofHint = floorTopicProofUILabel(latest)
		}
		watchlisted := q.ID == "Q.01"
		catStr := q.FloorCategoryLabel()
		browse = append(browse, map[string]any{
			"topic_id":               q.ID,
			"title":                  q.Title,
			"topic_class":            floorTopicsTopicClassPretty(catStr),
			"category":               floorTopicsBrowseCategoryFromQuestionCat(catStr),
			"probability_long":       pl,
			"probability_short":      ps,
			"probability_delta":      q.ProbabilityDelta,
			"consensus_status":       cs,
			"deadline_label":         floorTopicsDeadlineLabel(q.Deadline),
			"agent_count":            q.AgentCount,
			"top_signal_hint":        topHint,
			"proof_hint":             proofHint,
			"open_topic_details_url": "/topic/" + q.ID,
			"watchlisted":            watchlisted,
		})
	}
	if firstQ.ID == "" {
		return browse, nil, nil
	}
	cs0 := floorBrowseConsensusStatusFromQuestion(firstQ)
	tl := map[string]any{"agent_name": "—", "proof_label": nil}
	ts := map[string]any{"agent_name": "—", "proof_label": nil}
	if firstLong != nil && firstLong.Agent.Name != "" {
		tl["agent_name"] = firstLong.Agent.Name
		tl["proof_label"] = floorTopicProofUILabel(firstLong)
	}
	if firstShort != nil && firstShort.Agent.Name != "" {
		ts["agent_name"] = firstShort.Agent.Name
		ts["proof_label"] = floorTopicProofUILabel(firstShort)
	}
	fcat := firstQ.FloorCategoryLabel()
	sel := map[string]any{
		"topic_id":               firstQ.ID,
		"title":                  firstQ.Title,
		"topic_class":            floorTopicsTopicClassPretty(fcat),
		"probability_long":       firstQ.Probability,
		"probability_short":      1 - firstQ.Probability,
		"probability_delta":      firstQ.ProbabilityDelta,
		"consensus_status":       cs0,
		"participation_context":  floorParticipationContextFromQuestion(firstQ),
		"top_long_preview":       tl,
		"top_short_preview":      ts,
		"open_topic_details_url": "/topic/" + firstQ.ID,
		"open_research_url":      "/research",
	}
	chart := map[string]any{
		"kind": "donut", "long_percent": firstQ.Probability, "short_percent": 1 - firstQ.Probability,
	}
	return browse, sel, chart
}

func floorTopicsApplyQueryFilters(out map[string]any, r *http.Request) {
	cat := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("category")))
	if cat == "" {
		cat = "all"
	}
	raw, ok := out["browse_rows"]
	if !ok || raw == nil {
		return
	}
	rows, ok := raw.([]map[string]any)
	if !ok {
		return
	}
	filtered := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		if cat == "all" {
			filtered = append(filtered, row)
			continue
		}
		rv, _ := row["category"].(string)
		if strings.EqualFold(strings.TrimSpace(rv), cat) {
			filtered = append(filtered, row)
		}
	}
	if len(filtered) > 0 {
		out["browse_rows"] = filtered
	}
	if cats, ok := out["categories"].([]map[string]any); ok {
		for i := range cats {
			val, _ := cats[i]["value"].(string)
			cats[i]["active"] = strings.EqualFold(strings.TrimSpace(val), cat)
		}
	}
}

func (s *Server) handleFloorTopicsPage(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	out := floorComposedTopicsPage()
	var positions []dbpkg.FloorPosition
	if err := db.Preload("Agent").Preload("Question").Preload("Question.Category").Order("staked_at DESC, id DESC").Limit(50).Find(&positions).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	if len(positions) > 0 {
		browse, selTopic, selChart := floorTopicsBrowseBundleFromPositions(positions)
		if len(browse) > 0 {
			out["browse_rows"] = browse
			if selTopic != nil {
				out["selected_topic"] = selTopic
			}
			if selChart != nil {
				out["selected_topic_chart"] = selChart
			}
			if rr, ok := out["right_rail"].(map[string]any); ok {
				if reg, ok := rr["regional_divergence"].(map[string]any); ok && len(browse) > 0 {
					firstID := browse[0]["topic_id"]
					reg["summary"] = fmt.Sprintf("CN short vs US long on %v", firstID)
					reg["open_regional_detail_url"] = fmt.Sprintf("/topic/%v?view=regional&timeframe=7d", firstID)
				}
			}
		}
	}
	floorTopicsApplyQueryFilters(out, r)
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleFloorListCategories(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)
	var rows []dbpkg.Category
	if err := db.Where("is_active = ?", true).Order("sort_order ASC, id ASC").Find(&rows).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		c := &rows[i]
		out = append(out, map[string]any{
			"id":           c.ID,
			"display_name": c.DisplayName,
			"sort_order":   c.SortOrder,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"count": len(out), "items": out})
}

func (s *Server) mountFloorAPI(r chi.Router) {
	r.Route("/floor", func(fr chi.Router) {
		fr.Get("/categories", s.handleFloorListCategories)
		fr.Post("/topic-proposals", s.handleFloorCreateTopicProposal)
		fr.Get("/digests", s.handleFloorDigestStrip)
		fr.Get("/positions/{positionID}/challenges", s.handleFloorPositionChallenges)
		fr.Get("/positions", s.handleFloorGlobalPositions)
		fr.Get("/topics", s.handleFloorTopicsPage)
		fr.Get("/discover", s.handleFloorDiscoverPage)
		fr.Get("/index/{indexID}/detail", s.handleFloorIndexDetail)
		fr.Get("/index", s.handleFloorIndexPage)
		fr.Get("/topics/{questionID}/detail", s.handleFloorGetTopicDetails)
		fr.Get("/topics/{questionID}/regional", s.handleFloorGetTopicRegional)
		fr.Get("/topics/{questionID}/digest-history", s.handleFloorQuestionDigests)
		fr.Get("/questions/featured", s.handleFloorFeaturedQuestion)
		fr.Get("/questions", s.handleFloorListQuestions)
		fr.Get("/questions/{questionID}", s.handleFloorGetQuestion)
		fr.Get("/questions/{questionID}/context/worldmonitor", s.handleFloorQuestionWorldMonitorContext)
		fr.Get("/questions/{questionID}/positions", s.handleFloorQuestionPositions)
		fr.Get("/questions/{questionID}/digest-history", s.handleFloorQuestionDigests)
		fr.Get("/questions/{questionID}/digests", s.handleFloorQuestionDigests)
		fr.Get("/questions/{questionID}/probability-series", s.handleFloorProbabilitySeries)
		fr.Get("/agents/{agentID}/positions", s.handleFloorAgentPositions)
		fr.Get("/agents/{agentID}/topic-stats", s.handleFloorAgentTopicStats)
		fr.Get("/agents/{agentID}/signal-profile", s.handleFloorAgentSignalProfile)
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
	qry := db.Model(&dbpkg.FloorQuestion{}).Preload("Category")
	if status != "" {
		qry = qry.Where("status = ?", status)
	}
	if category != "" {
		qry = qry.Where("category_id = ?", category)
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
	err := db.Preload("Category").Order("staked_count DESC, agent_count DESC, id ASC").Limit(1).First(&q).Error
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
	if err := db.Preload("Category").First(&q, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeDetail(w, http.StatusNotFound, "Question not found")
			return
		}
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("view")), "regional") {
		writeJSON(w, http.StatusOK, floorComposedTopicRegionalPayload(&q, r, db))
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
	inferredCluster := strings.TrimSpace(r.URL.Query().Get("inferred_cluster"))
	speculative := strings.TrimSpace(r.URL.Query().Get("speculative"))
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
	if inferredCluster != "" {
		qry = qry.Where("LOWER(TRIM(inferred_cluster_at_stake)) = ?", strings.ToLower(inferredCluster))
	} else if cluster != "" {
		if floorQueryClusterIsInferredStyle(cluster) {
			qry = qry.Where("LOWER(TRIM(inferred_cluster_at_stake)) = ?", strings.ToLower(cluster))
		} else {
			qry = qry.Where("regional_cluster = ?", cluster)
		}
	}
	if speculative == "1" || strings.EqualFold(speculative, "true") {
		qry = qry.Where("speculative = ?", true)
	} else if speculative == "0" || strings.EqualFold(speculative, "false") {
		qry = qry.Where("speculative = ?", false)
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
	attachWM := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("include_external_signals")), "true") ||
		strings.TrimSpace(r.URL.Query().Get("include_external_signals")) == "1"
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		m := floorDigestMap(&rows[i])
		if attachWM {
			floorDigestAttachExternalSignals(db, m)
		}
		out = append(out, m)
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
	attachWM := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("include_external_signals")), "true") ||
		strings.TrimSpace(r.URL.Query().Get("include_external_signals")) == "1"
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		m := floorDigestMap(&rows[i])
		if attachWM {
			floorDigestAttachExternalSignals(db, m)
		}
		out = append(out, m)
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
	writeJSON(w, http.StatusOK, map[string]any{
		"agent_id":               agentID,
		"topic_stats":            statMaps,
		"inference":              infRow,
		"position_count":         totalPositions,
		"position_pending_count": pendingPositions,
	})
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
	if err := db.Order("edition_digest_date DESC NULLS LAST, is_featured DESC, list_sort ASC, id ASC").
		Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
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
