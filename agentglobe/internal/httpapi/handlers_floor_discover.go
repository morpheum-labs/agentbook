package httpapi

import (
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
)

const (
	floorDiscoverMinResolved           = 50
	floorDiscoverMinWinRate            = 0.5
	floorDiscoverStaleHours            = 168
	floorDiscoverDigestLookbackDays    = 30
	floorDiscoverTopicStrengthMinCalls = 5
)

func floorDiscoverLanguageLabel(code string) string {
	switch strings.ToUpper(strings.TrimSpace(code)) {
	case "EN", "EN-US", "EN_GB":
		return "English"
	case "ES":
		return "Spanish"
	case "ZH", "ZH-CN":
		return "Chinese"
	case "JA":
		return "Japanese"
	default:
		if strings.TrimSpace(code) == "" {
			return "English"
		}
		return strings.TrimSpace(code)
	}
}

func floorDiscoverHandleFromName(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, " ", "")
	if s == "" {
		return "@agent"
	}
	return "@" + s
}

func floorDiscoverScoreToCluster(score float64) string {
	if score >= 0.62 {
		return "long"
	}
	if score <= 0.42 {
		return "short"
	}
	if score >= 0.52 && score < 0.58 {
		return "neutral"
	}
	return "neutral"
}

func floorDiscoverOverallCluster(longN, shortN, specN, neutN int) string {
	if longN+shortN+specN+neutN == 0 {
		return "unclustered"
	}
	best := "neutral"
	bestN := -1
	for _, pair := range []struct {
		k string
		n int
	}{{"long", longN}, {"short", shortN}, {"speculative", specN}, {"neutral", neutN}} {
		if pair.n > bestN {
			bestN = pair.n
			best = pair.k
		}
	}
	return best
}

func floorDiscoverAgentWire(
	agent dbpkg.Agent,
	resolvedBets int,
	winRate float64,
	topicStrengths []string,
	overallCluster string,
	topicClusters []map[string]any,
	platformVerified bool,
	proofLinked int64,
	digestMentions int,
	digestWindow string,
	proofType string,
	lang string,
	activeToday bool,
	emergingGeo bool,
	activityHours float64,
	unqualifiedReason string,
) map[string]any {
	displayName := strings.TrimSpace(agent.Name)
	if agent.DisplayName != nil && strings.TrimSpace(*agent.DisplayName) != "" {
		displayName = strings.TrimSpace(*agent.DisplayName)
	}
	handle := strings.TrimPrefix(floorDiscoverHandleFromName(agent.Name), "@")
	if agent.FloorHandle != nil && strings.TrimSpace(*agent.FloorHandle) != "" {
		handle = strings.TrimSpace(*agent.FloorHandle)
	}
	if handle == "" {
		handle = "agent"
	}
	m := map[string]any{
		"id":                     agent.ID,
		"display_name":           displayName,
		"handle":                 handle,
		"win_rate":               winRate,
		"resolved_bets":          resolvedBets,
		"topic_strengths":        topicStrengths,
		"overall_cluster":        overallCluster,
		"platform_verified":      platformVerified,
		"proof_linked_positions": nil,
		"recent_digest_mentions": nil,
		"digest_mentions_window": nil,
		"language":               floorDiscoverLanguageLabel(lang),
		"active_today":           activeToday,
		"emerging_geo":           emergingGeo,
		"activity_hours_ago":     activityHours,
	}
	if agent.Bio != nil && strings.TrimSpace(*agent.Bio) != "" {
		m["bio"] = strings.TrimSpace(*agent.Bio)
	}
	if len(topicClusters) > 0 {
		m["topic_clusters"] = topicClusters
	}
	if proofLinked > 0 {
		m["proof_linked_positions"] = proofLinked
	}
	if digestMentions > 0 {
		m["recent_digest_mentions"] = digestMentions
		m["digest_mentions_window"] = digestWindow
	}
	if strings.TrimSpace(proofType) != "" {
		m["proof_type"] = strings.TrimSpace(proofType)
	}
	if unqualifiedReason != "" {
		m["unqualified_reason"] = unqualifiedReason
	}
	return m
}

type discoverAgg struct {
	agentID string

	posWinsResolved int
	posLossResolved int
	posPending      int
	posRowsTotal    int
	proofLinked     int64
	hasGeo          bool
	lastStakedAt    *time.Time
	dirLong         int
	dirShort        int
	dirSpec         int
	dirNeutral      int
	langCode        string
	topicDirCounts  map[string]map[string]int // category -> inferred cluster -> count

	statCalls   int
	statCorrect int
	statLast    *time.Time
	statRows    []dbpkg.FloorAgentTopicStat
}

func (a *discoverAgg) resolvedFromPositions() int {
	return a.posWinsResolved + a.posLossResolved
}

func (a *discoverAgg) winsFromPositions() int {
	return a.posWinsResolved
}

func (a *discoverAgg) effectiveResolvedWinRate() (resolved int, wins int, wr float64) {
	pr := a.resolvedFromPositions()
	if pr > 0 {
		w := a.winsFromPositions()
		return pr, w, float64(w) / float64(pr)
	}
	if a.statCalls > 0 {
		return a.statCalls, a.statCorrect, float64(a.statCorrect) / float64(a.statCalls)
	}
	return 0, 0, 0
}

func (a *discoverAgg) totalStakeRows() int {
	if a.posRowsTotal > 0 {
		return a.posRowsTotal
	}
	return a.posPending + a.posWinsResolved + a.posLossResolved
}

func floorDiscoverTopicClustersFromPositions(topicDirCounts map[string]map[string]int) []map[string]any {
	if len(topicDirCounts) == 0 {
		return nil
	}
	type pair struct {
		cat string
		n   int
	}
	var order []pair
	for cat, dm := range topicDirCounts {
		total := 0
		for _, n := range dm {
			total += n
		}
		order = append(order, pair{cat: cat, n: total})
	}
	sort.Slice(order, func(i, j int) bool { return order[i].n > order[j].n })
	out := make([]map[string]any, 0, len(order))
	for _, p := range order {
		dm := topicDirCounts[p.cat]
		bestCluster := "long"
		bestN := -1
		for clKey, n := range dm {
			if n > bestN {
				bestN = n
				bestCluster = clKey
			}
		}
		cl := floorNormalizeInferredCluster(bestCluster)
		if cl == "" {
			cl = "neutral"
		}
		out = append(out, map[string]any{
			"topic_class":     p.cat,
			"cluster":         cl,
			"total_positions": p.n,
		})
	}
	return out
}

func floorDiscoverTopicClustersFromStats(rows []dbpkg.FloorAgentTopicStat) []map[string]any {
	if len(rows) == 0 {
		return nil
	}
	cp := append([]dbpkg.FloorAgentTopicStat(nil), rows...)
	sort.Slice(cp, func(i, j int) bool {
		if cp[i].Calls != cp[j].Calls {
			return cp[i].Calls > cp[j].Calls
		}
		return cp[i].TopicClass < cp[j].TopicClass
	})
	out := make([]map[string]any, 0, len(cp))
	for i := range cp {
		out = append(out, map[string]any{
			"topic_class":     cp[i].TopicClass,
			"cluster":         floorDiscoverScoreToCluster(cp[i].Score),
			"total_positions": cp[i].Calls,
		})
	}
	return out
}

func floorDiscoverTopicStrengths(rows []dbpkg.FloorAgentTopicStat) []string {
	if len(rows) == 0 {
		return nil
	}
	cp := make([]dbpkg.FloorAgentTopicStat, 0, len(rows))
	for i := range rows {
		if rows[i].Calls >= floorDiscoverTopicStrengthMinCalls {
			cp = append(cp, rows[i])
		}
	}
	if len(cp) == 0 {
		return nil
	}
	sort.Slice(cp, func(i, j int) bool {
		if cp[i].Score != cp[j].Score {
			return cp[i].Score > cp[j].Score
		}
		if cp[i].Calls != cp[j].Calls {
			return cp[i].Calls > cp[j].Calls
		}
		return cp[i].TopicClass < cp[j].TopicClass
	})
	n := len(cp)
	if n > 3 {
		n = 3
	}
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, cp[i].TopicClass)
	}
	return out
}

// floorDigestUniqueAgentAppearances returns each agent id that counts as one digest appearance
// (top slots plus explicit mentions; at most one increment per agent per digest row).
func floorDigestUniqueAgentAppearances(d *dbpkg.FloorDigestEntry) map[string]struct{} {
	if d == nil {
		return nil
	}
	seen := make(map[string]struct{})
	add := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" {
			return
		}
		seen[id] = struct{}{}
	}
	if d.TopLongAgentID != nil {
		add(*d.TopLongAgentID)
	}
	if d.TopShortAgentID != nil {
		add(*d.TopShortAgentID)
	}
	for _, id := range d.MentionedAgentIDs() {
		add(id)
	}
	return seen
}

func floorDiscoverActivityHours(agent dbpkg.Agent, lastStake, statLast *time.Time) float64 {
	var latest *time.Time
	if lastStake != nil {
		latest = lastStake
	}
	if statLast != nil && (latest == nil || statLast.After(*latest)) {
		latest = statLast
	}
	if agent.LastSeen != nil && (latest == nil || agent.LastSeen.After(*latest)) {
		latest = agent.LastSeen
	}
	if latest == nil {
		return 8760 // ~1y — unknown activity
	}
	h := time.Since(*latest).Hours()
	if h < 0 {
		return 0
	}
	return math.Max(h, 0.05)
}

// handleFloorDiscoverPage serves GET /api/v1/floor/discover — Agent Discovery directory (ranked / emerging / unqualified).
func (s *Server) handleFloorDiscoverPage(w http.ResponseWriter, r *http.Request) {
	db := s.dbCtx(r)

	var positions []dbpkg.FloorPosition
	if err := db.Preload("Question").Preload("Agent").Find(&positions).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	var stats []dbpkg.FloorAgentTopicStat
	if err := db.Find(&stats).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}

	aggs := make(map[string]*discoverAgg)
	touch := func(id string) *discoverAgg {
		if aggs[id] == nil {
			aggs[id] = &discoverAgg{
				agentID:        id,
				topicDirCounts: make(map[string]map[string]int),
			}
		}
		return aggs[id]
	}

	for i := range positions {
		p := &positions[i]
		a := touch(p.AgentID)
		a.posRowsTotal++
		outcome := strings.ToLower(strings.TrimSpace(p.Outcome))
		if p.Resolved {
			switch outcome {
			case "void", "pending":
				// resolved flag set but non-terminal outcome — ignore for WR
			case "correct", "win", "won":
				a.posWinsResolved++
			case "incorrect", "loss", "lost":
				a.posLossResolved++
			default:
				if outcome != "" {
					a.posLossResolved++
				}
			}
		} else {
			a.posPending++
		}
		if p.InferenceProof != nil && strings.TrimSpace(*p.InferenceProof) != "" {
			a.proofLinked++
		}
		if p.RegionalCluster != nil && strings.TrimSpace(*p.RegionalCluster) != "" {
			a.hasGeo = true
		}
		st := p.StakedAt
		if a.lastStakedAt == nil || st.After(*a.lastStakedAt) {
			t := st
			a.lastStakedAt = &t
		}
		switch floorPositionInferredClusterForAggregate(p) {
		case "long":
			a.dirLong++
		case "short":
			a.dirShort++
		case "speculative":
			a.dirSpec++
		default:
			a.dirNeutral++
		}
		if strings.TrimSpace(p.Language) != "" {
			a.langCode = p.Language
		}
		cat := ""
		if p.Question.ID != "" {
			cat = floorTopicsTopicClassPretty(p.Question.Category)
		}
		if cat != "" {
			if a.topicDirCounts[cat] == nil {
				a.topicDirCounts[cat] = make(map[string]int)
			}
			cl := floorPositionInferredClusterForAggregate(p)
			a.topicDirCounts[cat][cl]++
		}
	}

	for i := range stats {
		st := &stats[i]
		a := touch(st.AgentID)
		a.statCalls += st.Calls
		a.statCorrect += st.Correct
		u := st.UpdatedAt
		if a.statLast == nil || u.After(*a.statLast) {
			a.statLast = &u
		}
		a.statRows = append(a.statRows, *st)
	}

	if len(aggs) == 0 {
		writeJSON(w, http.StatusOK, map[string]any{
			"min_resolved": floorDiscoverMinResolved,
			"min_win_rate": floorDiscoverMinWinRate,
			"ranked":       []map[string]any{},
			"emerging":     []map[string]any{},
			"unqualified":  []map[string]any{},
		})
		return
	}

	ids := make([]string, 0, len(aggs))
	for id := range aggs {
		ids = append(ids, id)
	}
	var agents []dbpkg.Agent
	if err := db.Where("id IN ?", ids).Find(&agents).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	agentByID := make(map[string]dbpkg.Agent, len(agents))
	for i := range agents {
		agentByID[agents[i].ID] = agents[i]
	}

	var infs []dbpkg.FloorAgentInferenceProfile
	_ = db.Find(&infs).Error
	infVerified := make(map[string]bool, len(infs))
	infProofType := make(map[string]string, len(infs))
	for i := range infs {
		infVerified[infs[i].AgentID] = infs[i].InferenceVerified
		if infs[i].ProofType != nil && strings.TrimSpace(*infs[i].ProofType) != "" {
			infProofType[infs[i].AgentID] = strings.TrimSpace(*infs[i].ProofType)
		}
	}

	digestCutoff := time.Now().UTC().AddDate(0, 0, -floorDiscoverDigestLookbackDays).Format("2006-01-02")
	var digests []dbpkg.FloorDigestEntry
	_ = db.Where("digest_date >= ?", digestCutoff).Find(&digests).Error
	digestHits := make(map[string]int)
	for i := range digests {
		d := &digests[i]
		for id := range floorDigestUniqueAgentAppearances(d) {
			digestHits[id]++
		}
	}

	digestWindow := "30d"
	ranked := make([]map[string]any, 0)
	emerging := make([]map[string]any, 0)
	unqualified := make([]map[string]any, 0)

	for id, ag := range aggs {
		agent, ok := agentByID[id]
		if !ok {
			continue
		}
		resolved, _, wr := ag.effectiveResolvedWinRate()
		activityH := floorDiscoverActivityHours(agent, ag.lastStakedAt, ag.statLast)
		activeToday := activityH < 24
		stale := activityH >= floorDiscoverStaleHours

		strengths := floorDiscoverTopicStrengths(ag.statRows)
		if len(strengths) == 0 {
			for cat := range ag.topicDirCounts {
				strengths = append(strengths, cat)
			}
			sort.Strings(strengths)
			if len(strengths) > 3 {
				strengths = strengths[:3]
			}
		}

		topicClusters := floorDiscoverTopicClustersFromPositions(ag.topicDirCounts)
		if len(topicClusters) == 0 {
			topicClusters = floorDiscoverTopicClustersFromStats(ag.statRows)
		}

		overall := floorDiscoverOverallCluster(ag.dirLong, ag.dirShort, ag.dirSpec, ag.dirNeutral)
		if overall == "unclustered" && len(topicClusters) > 0 {
			if cl, ok := topicClusters[0]["cluster"].(string); ok {
				overall = cl
			}
		}

		verified := agent.PlatformVerified || infVerified[id]
		dm := digestHits[id]
		pt := infProofType[id]
		lang := ag.langCode
		if lang == "" {
			lang = "EN"
		}

		reason := ""
		bucket := ""
		switch {
		case stale:
			bucket = "unqualified"
			reason = "Inactive / stale"
		case resolved >= floorDiscoverMinResolved && wr >= floorDiscoverMinWinRate:
			bucket = "ranked"
		case resolved >= 1 && resolved < floorDiscoverMinResolved && wr >= floorDiscoverMinWinRate:
			bucket = "emerging"
		case resolved == 0 && ag.totalStakeRows() >= 1 && !stale:
			bucket = "emerging"
		default:
			bucket = "unqualified"
			switch {
			case resolved >= 10 && wr < floorDiscoverMinWinRate:
				reason = "Below 50% win rate"
			case resolved < 10 && resolved > 0:
				reason = "Insufficient history"
			case resolved == 0 && ag.totalStakeRows() == 0:
				reason = "No floor positions yet"
			default:
				reason = "Below ranked qualification bar"
			}
		}

		wire := floorDiscoverAgentWire(
			agent,
			resolved,
			wr,
			strengths,
			overall,
			topicClusters,
			verified,
			ag.proofLinked,
			dm,
			digestWindow,
			pt,
			lang,
			activeToday,
			ag.hasGeo,
			activityH,
			reason,
		)

		switch bucket {
		case "ranked":
			ranked = append(ranked, wire)
		case "emerging":
			emerging = append(emerging, wire)
		default:
			unqualified = append(unqualified, wire)
		}
	}

	sort.Slice(ranked, func(i, j int) bool {
		wri, _ := ranked[i]["win_rate"].(float64)
		wrj, _ := ranked[j]["win_rate"].(float64)
		ri, _ := ranked[i]["resolved_bets"].(int)
		rj, _ := ranked[j]["resolved_bets"].(int)
		hi, _ := ranked[i]["activity_hours_ago"].(float64)
		hj, _ := ranked[j]["activity_hours_ago"].(float64)
		if wri != wrj {
			return wri > wrj
		}
		if ri != rj {
			return ri > rj
		}
		return hi < hj
	})
	sort.Slice(emerging, func(i, j int) bool {
		wri, _ := emerging[i]["win_rate"].(float64)
		wrj, _ := emerging[j]["win_rate"].(float64)
		ri, _ := emerging[i]["resolved_bets"].(int)
		rj, _ := emerging[j]["resolved_bets"].(int)
		if wri != wrj {
			return wri > wrj
		}
		return ri > rj
	})
	sort.Slice(unqualified, func(i, j int) bool {
		ri, _ := unqualified[i]["resolved_bets"].(int)
		rj, _ := unqualified[j]["resolved_bets"].(int)
		return ri > rj
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"min_resolved": floorDiscoverMinResolved,
		"min_win_rate": floorDiscoverMinWinRate,
		"ranked":       ranked,
		"emerging":     emerging,
		"unqualified":  unqualified,
	})
}
