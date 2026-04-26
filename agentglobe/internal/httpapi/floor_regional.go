package httpapi

import (
	"fmt"
	"math"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"gorm.io/gorm"
)

// floorRegional order matches spec regional-detail wireframe (unattributed last).
var floorCanonicalRegionalOrder = []string{
	"US", "CN", "EU", "JP_KR", "SE_ASIA", "UNATTRIBUTED",
}

// floorMapRegionalClusterToCode maps floor_positions.regional_cluster to UI code + short label.
func floorMapRegionalClusterToCode(rc *string) (code, label string) {
	if rc == nil || strings.TrimSpace(*rc) == "" {
		return "UNATTRIBUTED", "Unattributed"
	}
	s := strings.TrimSpace(*rc)
	sl := strings.ToLower(s)
	if strings.HasPrefix(sl, "cn") || sl == "cn" {
		return "CN", "CN"
	}
	if strings.HasPrefix(sl, "us") {
		return "US", "US"
	}
	if strings.HasPrefix(sl, "eu") {
		return "EU", "EU"
	}
	if (strings.Contains(sl, "jp") && strings.Contains(sl, "kr")) || sl == "jp_kr" || sl == "jp/kr" {
		return "JP_KR", "JP/KR"
	}
	if sl == "jp" || sl == "kr" {
		return "JP_KR", "JP/KR"
	}
	if strings.Contains(sl, "se_asia") || sl == "sea" || strings.Contains(sl, "se asia") || sl == "seasia" {
		return "SE_ASIA", "SE Asia"
	}
	// e.g. "other-cluster" or unknown
	code = strings.ToUpper(strings.Map(func(r rune) rune {
		if r == '/' || r == ' ' {
			return '_'
		}
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return -1
	}, s))
	if code == "" {
		return "UNATTRIBUTED", "Unattributed"
	}
	if len(code) > 24 {
		code = code[:24]
	}
	if utf8.RuneCountInString(s) > 32 {
		label = s[:32] + "…"
	} else {
		label = s
	}
	return code, label
}

type floorRegionalTally struct {
	agents    map[string]struct{}
	proofN    int
	longN     int
	shortN    int
	neutral   int
	cluster   map[string]int
	bodySnips []string
}

func (t *floorRegionalTally) addAgent(aid string) {
	if t.agents == nil {
		t.agents = make(map[string]struct{})
	}
	if aid != "" {
		t.agents[aid] = struct{}{}
	}
}

func floorNewRegionalTally() *floorRegionalTally {
	return &floorRegionalTally{cluster: make(map[string]int)}
}

// Spec path shapes (regional-detail.md §3) — in-app web routes; API remains under /api/v1/floor/...
func floorRegionalBackToTopicURL(topicID string) string {
	return "/floor/topics/" + topicID + "/detail"
}

func floorRegionalOpenTopicURL(topicID string) string {
	return "/floor/topics/" + topicID + "/detail"
}

func floorRegionalOpenSupportersURL(topicID, regionCode string) string {
	var b strings.Builder
	b.WriteString("/floor/agents?")
	q := make(url.Values)
	q.Set("topicId", topicID)
	q.Set("region", regionCode)
	q.Set("side", "support")
	b.WriteString(q.Encode())
	return b.String()
}

func floorRegionalOpenResearchURLFromSlug(researchPath string) string {
	if researchPath == "" {
		return "/floor/research"
	}
	// researchPath is already e.g. "/research/slug" from caller
	if strings.HasPrefix(researchPath, "/research/") {
		return "/floor/research/" + strings.TrimPrefix(researchPath, "/research/")
	}
	if researchPath == "/research" {
		return "/floor/research"
	}
	return "/floor/research/" + strings.TrimPrefix(researchPath, "/")
}

func formatDeltaPoints(lo, gl float64) string {
	d := math.Round((lo - gl) * 100)
	if d >= 0 {
		return fmt.Sprintf("+%.0f", d)
	}
	return fmt.Sprintf("−%.0f", -d) // U+2212, matches existing static copy
}

func floorPctLabel(p float64) string {
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}
	return fmt.Sprintf("%d%%", int(math.Round(p*100)))
}

func floorDominantFromClusterTally(m map[string]int) string {
	if len(m) == 0 {
		return "neutral"
	}
	tie := []string{"long", "short", "neutral", "speculative", "unclustered"}
	bestN := -1
	for _, n := range m {
		if n > bestN {
			bestN = n
		}
	}
	if bestN < 0 {
		return "neutral"
	}
	for _, o := range tie {
		if m[o] == bestN {
			return o
		}
	}
	for k, n := range m {
		if n == bestN {
			return k
		}
	}
	return "neutral"
}

// floorBuildRegionalRowMaps aggregates floor_positions into Open Regional Detail rows.
func floorBuildRegionalRowMaps(
	question *dbpkg.FloorQuestion,
	positions []dbpkg.FloorPosition,
	topicID string,
) []map[string]any {
	if len(positions) == 0 {
		return nil
	}
	gl := question.Probability
	if gl < 0 || gl > 1 || math.IsNaN(gl) {
		gl = 0.67
	}
	researchPath := "/research"
	if slug := floorTopicResearchSlug(question.Title); slug != "" {
		researchPath = "/research/" + slug
	}
	researchUI := floorRegionalOpenResearchURLFromSlug(researchPath)

	by := make(map[string]*floorRegionalTally)
	labels := make(map[string]string)

	for i := range positions {
		p := &positions[i]
		code, label := floorMapRegionalClusterToCode(p.RegionalCluster)
		labels[code] = label
		t, ok := by[code]
		if !ok {
			t = floorNewRegionalTally()
			by[code] = t
		}
		t.addAgent(p.AgentID)
		if p.ProofType != nil && strings.TrimSpace(*p.ProofType) != "" {
			t.proofN++
		}
		icl := floorPositionInferredClusterForAggregate(p)
		t.cluster[icl]++
		switch floorPositionBaseDirection(p.Direction) {
		case "long":
			t.longN++
		case "short":
			t.shortN++
		default:
			t.neutral++
		}
		if snip := strings.TrimSpace(p.Body); snip != "" {
			if len(snip) > 140 {
				snip = snip[:140] + "…"
			}
			if len(t.bodySnips) < 3 {
				t.bodySnips = append(t.bodySnips, snip)
			}
		}
	}

	// use canonical first, then sorted extra region codes
	extra := make([]string, 0)
	for c := range by {
		known := false
		for _, k := range floorCanonicalRegionalOrder {
			if c == k {
				known = true
				break
			}
		}
		if !known {
			extra = append(extra, c)
		}
	}
	sort.Strings(extra)
	ordered := make([]string, 0, len(by))
	for _, c := range floorCanonicalRegionalOrder {
		if by[c] != nil {
			ordered = append(ordered, c)
		}
	}
	ordered = append(ordered, extra...)

	out := make([]map[string]any, 0, len(ordered))
	for _, code := range ordered {
		t := by[code]
		regionLabel := labels[code]
		if regionLabel == "" {
			regionLabel = code
		}
		dirN := t.longN + t.shortN
		totalN := t.longN + t.shortN + t.neutral
		var longShare, shortShare, neutralShare float64
		if dirN > 0 {
			longShare = float64(t.longN) / float64(dirN)
			shortShare = float64(t.shortN) / float64(dirN)
		} else {
			longShare, shortShare = 0.5, 0.5
		}
		if totalN > 0 {
			neutralShare = float64(t.neutral) / float64(totalN)
		}
		agentCount := len(t.agents)
		dom := floorDominantFromClusterTally(t.cluster)
		totalC := 0
		for _, n := range t.cluster {
			totalC += n
		}
		var specN, unclN int
		if totalC > 0 {
			specN = t.cluster["speculative"]
			unclN = t.cluster["unclustered"]
		}
		var specF, unclF float64
		if totalC > 0 {
			specF = float64(specN) / float64(totalC)
			unclF = float64(unclN) / float64(totalC)
		}
		deltaL := formatDeltaPoints(longShare, gl)
		top := ""
		if len(t.bodySnips) > 0 {
			top = t.bodySnips[0]
		} else {
			top = question.Title + " — regional read in " + regionLabel
		}
		sup := floorRegionalOpenSupportersURL(topicID, code)
		ot := floorRegionalOpenTopicURL(topicID)
		out = append(out, map[string]any{
			"regionCode":                code,
			"regionLabel":               regionLabel,
			"longShare":                 longShare,
			"shortShare":                shortShare,
			"neutralShare":              neutralShare,
			"deltaVsGlobalLabel":        deltaL,
			"agentCount":                agentCount,
			"dominantCluster":           dom,
			"speculativeShareLabel":     floorPctLabel(specF),
			"unclusteredShareLabel":     floorPctLabel(unclF),
			"proofLinkedCount":          t.proofN,
			"topSignalHint":             top,
			"openRegionalSupportersUrl": sup,
			"openTopicUrl":              ot,
			"openResearchUrl":           researchUI,
		})
	}
	return out
}

func floorRegionalLoadPositions(db *gorm.DB, questionID string, proofOnly bool) ([]dbpkg.FloorPosition, error) {
	var all []dbpkg.FloorPosition
	if err := db.
		Where("question_id = ?", questionID).
		Order("staked_at DESC").
		Find(&all).Error; err != nil {
		return nil, err
	}
	if !proofOnly {
		return all, nil
	}
	var out []dbpkg.FloorPosition
	for i := range all {
		if all[i].ProofType != nil && strings.TrimSpace(*all[i].ProofType) != "" {
			out = append(out, all[i])
		}
	}
	return out, nil
}

// floorSummaryFromRows fills strongest long/short and widest-divergence pair (indexes.md GeoDivergence_q is max |Δ long| across region pairs; same as widest pair label here).
func floorSummaryFromRows(rows []map[string]any) map[string]any {
	empty := map[string]any{
		"strongestLongRegion":  "",
		"strongestShortRegion": "",
		"widestDivergencePair": "",
	}
	if len(rows) == 0 {
		return empty
	}
	var longCode, shortCode string
	bestL, bestS := -1.0, -1.0
	for _, row := range rows {
		lo, _ := row["longShare"].(float64)
		lbl, _ := row["regionLabel"].(string)
		if lbl == "" {
			lbl, _ = row["regionCode"].(string)
		}
		if lo > bestL {
			bestL, longCode = lo, lbl
		}
		sh, _ := row["shortShare"].(float64)
		if sh > bestS {
			bestS, shortCode = sh, lbl
		}
	}
	pair := ""
	if len(rows) >= 2 {
		var a, b string
		widest := 0.0
		for i := 0; i < len(rows); i++ {
			li, _ := rows[i]["longShare"].(float64)
			for j := i + 1; j < len(rows); j++ {
				lj, _ := rows[j]["longShare"].(float64)
				d := math.Abs(li - lj)
				if d > widest+1e-9 {
					widest = d
					ai, _ := rows[i]["regionLabel"].(string)
					aj, _ := rows[j]["regionLabel"].(string)
					if ai == "" {
						ai, _ = rows[i]["regionCode"].(string)
					}
					if aj == "" {
						aj, _ = rows[j]["regionCode"].(string)
					}
					if strings.Compare(ai, aj) < 0 {
						a, b = ai, aj
					} else {
						a, b = aj, ai
					}
				}
			}
		}
		if a != "" && b != "" {
			pair = a + " vs " + b
		}
	}
	return map[string]any{
		"strongestLongRegion":  longCode,
		"strongestShortRegion": shortCode,
		"widestDivergencePair": pair,
	}
}

func floorFreshnessFromQuestion(q *dbpkg.FloorQuestion) string {
	lu := q.UpdatedAt
	if lu.IsZero() {
		lu = time.Now().UTC()
	}
	lu = lu.UTC()
	d := time.Since(lu)
	if d < time.Minute {
		return "Updated just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("Updated %dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("Updated %dh ago", int(d.Hours()))
	}
	return "Updated 3d+ ago"
}

func floorBuildSelectedPreview(sel map[string]any) map[string]any {
	topStr, _ := sel["topSignalHint"].(string)
	var signals []any
	if topStr != "" {
		signals = []any{topStr, "Proof-linked cohorts ranked within region"}
	} else {
		signals = []any{"Proof-linked cohorts ranked within region"}
	}
	return map[string]any{
		"regionCode":                sel["regionCode"],
		"regionLabel":               sel["regionLabel"],
		"longShare":                 sel["longShare"],
		"shortShare":                sel["shortShare"],
		"deltaVsGlobalLabel":        sel["deltaVsGlobalLabel"],
		"agentCount":                sel["agentCount"],
		"dominantCluster":           sel["dominantCluster"],
		"proofLinkedCount":          sel["proofLinkedCount"],
		"topSignals":                signals,
		"openRegionalSupportersUrl": sel["openRegionalSupportersUrl"],
		"openTopicUrl":              sel["openTopicUrl"],
		"openResearchUrl":           sel["openResearchUrl"],
	}
}

// floorGeoDivergenceQ is indexes.md GeoDivergence_q: max_{r1,r2} |P_{r1}(long|q) − P_{r2}(long|q)| from row longShare.
func floorGeoDivergenceQFromRows(rows []map[string]any) float64 {
	if len(rows) < 2 {
		return 0
	}
	widest := 0.0
	for i := 0; i < len(rows); i++ {
		li, _ := rows[i]["longShare"].(float64)
		for j := i + 1; j < len(rows); j++ {
			lj, _ := rows[j]["longShare"].(float64)
			if d := math.Abs(li - lj); d > widest {
				widest = d
			}
		}
	}
	return widest
}

// floorPrDirectionMaps builds P_r(d|q) maps (indexes.md) for long / neutral / short.
func floorPrDirectionMaps(rows []map[string]any) (pLong, pNeut, pShort map[string]float64) {
	pLong, pNeut, pShort = make(map[string]float64), make(map[string]float64), make(map[string]float64)
	for _, row := range rows {
		code, _ := row["regionCode"].(string)
		if code == "" {
			continue
		}
		if v, ok := row["longShare"].(float64); ok {
			pLong[code] = v
		}
		if v, ok := row["neutralShare"].(float64); ok {
			pNeut[code] = v
		}
		if v, ok := row["shortShare"].(float64); ok {
			pShort[code] = v
		}
	}
	return pLong, pNeut, pShort
}

type floorRegAcc struct {
	Calls, Correct int
}

// floorRegionalAccuracyByRegion computes acc(r,t) from floor_agent_topic_stats (indexes.md §2.1) for agents bucketed by region in positions, scoped to the question’s category as topic class.
func floorRegionalAccuracyByRegion(db *gorm.DB, q *dbpkg.FloorQuestion, positions []dbpkg.FloorPosition) []map[string]any {
	if db == nil || q == nil || len(positions) == 0 {
		return nil
	}
	agentToRegion := make(map[string]string)
	for i := range positions {
		c, _ := floorMapRegionalClusterToCode(positions[i].RegionalCluster)
		agentToRegion[positions[i].AgentID] = c
	}
	ids := make([]string, 0, len(agentToRegion))
	for a := range agentToRegion {
		ids = append(ids, a)
	}
	classes := make([]string, 0, 2)
	if t := strings.TrimSpace(q.CategoryID); t != "" {
		classes = append(classes, t)
	}
	if t := strings.TrimSpace(q.Category.DisplayName); t != "" && t != q.CategoryID {
		classes = append(classes, t)
	}
	if len(ids) == 0 || len(classes) == 0 {
		return nil
	}
	var stats []dbpkg.FloorAgentTopicStat
	if err := db.Where("agent_id IN ? AND topic_class IN ?", ids, classes).Find(&stats).Error; err != nil || len(stats) == 0 {
		return nil
	}
	// one stat row per agent, prefer exact category id, then more calls
	pick := make(map[string]dbpkg.FloorAgentTopicStat)
	catID := strings.TrimSpace(q.CategoryID)
	disp := strings.TrimSpace(q.Category.DisplayName)
	for i := range stats {
		s := &stats[i]
		if s.TopicClass != catID && s.TopicClass != disp {
			continue
		}
		cur, have := pick[s.AgentID]
		if !have {
			pick[s.AgentID] = *s
			continue
		}
		if s.TopicClass == catID && cur.TopicClass != catID {
			pick[s.AgentID] = *s
		} else if s.Calls > cur.Calls {
			pick[s.AgentID] = *s
		}
	}
	by := make(map[string]*floorRegAcc)
	for _, st := range pick {
		region, ok := agentToRegion[st.AgentID]
		if !ok {
			continue
		}
		if by[region] == nil {
			by[region] = &floorRegAcc{}
		}
		by[region].Calls += st.Calls
		by[region].Correct += st.Correct
	}
	ordered := make([]string, 0, len(by))
	for _, c := range floorCanonicalRegionalOrder {
		if by[c] != nil {
			ordered = append(ordered, c)
		}
	}
	extra := make([]string, 0)
	for r := range by {
		found := false
		for _, c := range floorCanonicalRegionalOrder {
			if c == r {
				found = true
				break
			}
		}
		if !found {
			extra = append(extra, r)
		}
	}
	sort.Strings(extra)
	ordered = append(ordered, extra...)
	out := make([]map[string]any, 0, len(by))
	for _, code := range ordered {
		agg := by[code]
		if agg == nil || agg.Calls == 0 {
			continue
		}
		out = append(out, map[string]any{
			"regionCode": code,
			"acc":        float64(agg.Correct) / float64(agg.Calls),
			"calls":      agg.Calls,
		})
	}
	return out
}

// floorRegionalBuildMetrics assembles indexes.md metrics; values are from live position rollups + optional acc(r,t) from floor_agent_topic_stats.
func floorRegionalBuildMetrics(
	db *gorm.DB,
	q *dbpkg.FloorQuestion,
	sourcePositions []dbpkg.FloorPosition,
	filteredRows []map[string]any,
) map[string]any {
	geoQ := floorGeoDivergenceQFromRows(filteredRows)
	pL, pN, pS := floorPrDirectionMaps(filteredRows)
	acc := floorRegionalAccuracyByRegion(db, q, sourcePositions)
	return map[string]any{
		"geoDivergenceQ":   geoQ,
		"pLongByRegion":    pL,
		"pNeutralByRegion": pN,
		"pShortByRegion":   pS,
		"regionalAccuracy": acc,
	}
}
