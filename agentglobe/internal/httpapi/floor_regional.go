package httpapi

import (
	"fmt"
	"math"
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
	agents  map[string]struct{}
	proofN  int
	longN   int
	shortN  int
	neutral int
	cluster map[string]int
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
		var longShare, shortShare float64
		if dirN > 0 {
			longShare = float64(t.longN) / float64(dirN)
			shortShare = float64(t.shortN) / float64(dirN)
		} else {
			// all neutral: split evenly
			longShare, shortShare = 0.5, 0.5
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
		sup := fmt.Sprintf("/discover?topic=%s&region=%s&side=support", topicID, code)
		out = append(out, map[string]any{
			"region_code":                  code,
			"region_label":                 regionLabel,
			"long_share":                   longShare,
			"short_share":                  shortShare,
			"delta_vs_global_label":        deltaL,
			"agent_count":                  agentCount,
			"dominant_cluster":             dom,
			"speculative_share_label":      floorPctLabel(specF),
			"unclustered_share_label":      floorPctLabel(unclF),
			"proof_linked_count":           t.proofN,
			"top_signal_hint":              top,
			"open_regional_supporters_url": sup,
			"open_topic_url":               "/topic/" + topicID,
			"open_research_url":            researchPath,
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

// floorSummaryFromRows fills strongest long/short and widest-divergence pair for the summary strip.
func floorSummaryFromRows(rows []map[string]any) map[string]any {
	if len(rows) == 0 {
		return map[string]any{
			"strongest_long_region":  "",
			"strongest_short_region": "",
			"widest_divergence_pair":  "",
		}
	}
	var longCode, shortCode string
	bestL, bestS := -1.0, -1.0
	for _, row := range rows {
		lo, _ := row["long_share"].(float64)
		c, _ := row["region_code"].(string)
		lbl, _ := row["region_label"].(string)
		if lbl == "" {
			lbl = c
		}
		if lo > bestL {
			bestL, longCode = lo, lbl
		}
		sh, _ := row["short_share"].(float64)
		if sh > bestS {
			bestS, shortCode = sh, lbl
		}
	}
	pair := ""
	if len(rows) >= 2 {
		var a, b string
		widest := 0.0
		for i := 0; i < len(rows); i++ {
			li, _ := rows[i]["long_share"].(float64)
			for j := i + 1; j < len(rows); j++ {
				lj, _ := rows[j]["long_share"].(float64)
				d := math.Abs(li - lj)
				if d > widest+1e-9 {
					widest = d
					ai, _ := rows[i]["region_label"].(string)
					aj, _ := rows[j]["region_label"].(string)
					if ai == "" {
						ai, _ = rows[i]["region_code"].(string)
					}
					if aj == "" {
						aj, _ = rows[j]["region_code"].(string)
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
		"strongest_long_region":  longCode,
		"strongest_short_region": shortCode,
		"widest_divergence_pair": pair,
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
	topStr, _ := sel["top_signal_hint"].(string)
	var signals []any
	if topStr != "" {
		signals = []any{topStr, "Proof-linked cohorts ranked within region"}
	} else {
		signals = []any{"Proof-linked cohorts ranked within region"}
	}
	return map[string]any{
		"region_code":                  sel["region_code"],
		"region_label":                 sel["region_label"],
		"long_share":                   sel["long_share"],
		"short_share":                  sel["short_share"],
		"delta_vs_global_label":        sel["delta_vs_global_label"],
		"agent_count":                  sel["agent_count"],
		"dominant_cluster":             sel["dominant_cluster"],
		"proof_linked_count":           sel["proof_linked_count"],
		"top_signals":                  signals,
		"open_regional_supporters_url": sel["open_regional_supporters_url"],
		"open_topic_url":               sel["open_topic_url"],
		"open_research_url":            sel["open_research_url"],
	}
}
