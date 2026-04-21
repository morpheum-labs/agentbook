package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
)

func floorMacroStripDefault() []map[string]any {
	return []map[string]any{
		{"label": "DXY", "value": "+0.3", "direction": "up"},
		{"label": "10Y", "value": "+8bp", "direction": "up"},
		{"label": "Oil", "value": "-1.2", "direction": "down"},
		{"label": "Gold", "value": "+0.7", "direction": "up"},
		{"label": "BTC", "value": "+2.1", "direction": "up"},
		{"label": "Liquidity", "value": "improving", "direction": "neutral"},
		{"label": "Vol", "value": "moderate", "direction": "neutral"},
	}
}

func floorComposedIndexDetailPageOnly(indexID string, watchlistLocked bool) map[string]any {
	id := strings.TrimSpace(indexID)
	type meta struct {
		title, subtitle, typeLabel, reading, why, access string
		conf                                               int
		topics                                             int
	}
	var m meta
	switch id {
	case "I.01":
		m = meta{"Retail Parking Lot Index", "VQ-Native · satellite retail flow", "VQ-Native", "Bullish divergence",
			"Leads retail earnings by weeks.", "premium", 82, 18}
	case "I.02":
		m = meta{"China Crematorium Activity Index", "Hidden Data · alternative stress", "Hidden Data", "High alert",
			"Non-traditional macro stress signal.", "premium", 84, 14}
	case "I.03":
		m = meta{"Truck Traffic Index", "Real-Time · freight demand", "Real-Time", "Softening WoW",
			"Freight pulse for goods demand.", "api", 71, 12}
	case "I.04":
		m = meta{"MAG7-style Basket", "SSI-Type · concentration lens", "SSI-Type", "Bullish drift MTD",
			"Concentration + rebalance risk in one lens.", "executable", 68, 9}
	case "I.00":
		m = meta{"Global Liquidity Pulse", "Macro · broad risk gauge", "Macro", "Neutral",
			"Broad risk-on / risk-off pressure gauge.", "free", 62, 22}
	default:
		return nil
	}
	openResearch := "/research"
	openFloor := "/"
	openDiscover := "/discover?supporting=true&indexId=" + id
	topic := func(qid, title, weight, score, contrib, mix, fresh string) map[string]any {
		return map[string]any{
			"topic_id": qid, "topic_title": title,
			"weight_label": weight, "topic_score_label": score, "contribution_label": contrib,
			"cluster_mix_label": mix, "freshness_label": fresh,
			"open_topic_url":      "/topic/" + qid,
			"open_research_url":   openResearch,
			"open_supporters_url": openDiscover + "&topicId=" + qid,
		}
	}
	bullet := func(text, qid string) map[string]any {
		out := map[string]any{"text": text, "open_research_url": openResearch}
		if qid != "" {
			out["open_topic_url"] = "/topic/" + qid
		}
		return out
	}
	return map[string]any{
		"index_id": id,
		"header": map[string]any{
			"title": m.title, "subtitle": m.subtitle, "type_label": m.typeLabel,
			"access_tier": m.access, "timeframe": "7d",
			"can_watchlist": true, "watchlist_locked": watchlistLocked,
		},
		"hero": map[string]any{
			"thesis":                  "Derived from " + strconv.Itoa(m.topics) + " topic results, weighted by cluster accuracy and topic relevance.",
			"current_reading":         m.reading,
			"why_it_matters_now":      m.why,
			"confidence_score":        m.conf,
			"freshness_label":         "Updated 5m ago",
			"topic_count":             m.topics,
			"unclustered_share_label": "11%",
			"method_label":            "Cluster-weighted topic index · speculative discount on · recompute every 15m",
			"open_floor_url":          openFloor,
			"open_research_url":       openResearch,
		},
		"macro_strip": floorMacroStripDefault(),
		"current_reading_body": "Reading is supported by " + strconv.Itoa(minInt(5, m.topics)) + " of top weighted topics; cluster-weighted blend.",
		"what_moved": []map[string]any{
			bullet("+ Leading flow topic strengthened", "Q.01"),
			bullet("+ Liquidity proxy topic improved", "Q.02"),
			bullet("- Policy uncertainty topic weighed", "Q.05"),
		},
		"topic_contribution_rows": []map[string]any{
			topic("Q.01", "Celtics will win the NBA Finals", "18%", "+0.28", "+0.050", "L-heavy", "3m"),
			topic("Q.02", "Fed rate cut — June meeting", "16%", "+0.22", "+0.035", "L/S mix", "5m"),
			topic("Q.03", "GPT-6 release before Q3 2026", "14%", "-0.10", "-0.014", "mixed", "22m"),
		},
		"counter_evidence": map[string]any{
			"severity_label": "Weak, but present",
			"items": []map[string]any{
				bullet("Topic breadth still concentrated in a few drivers", "Q.04"),
				bullet("One macro topic could flip momentum if data revises", "Q.02"),
			},
		},
		"signals_to_watch": []map[string]any{
			{"text": "Top bullish topics losing majority", "open_topic_url": "/topic/Q.01"},
			{"text": "Speculative share rising above threshold", "open_floor_url": openFloor},
			{"text": "Macro topic turning net short", "open_topic_url": "/topic/Q.02", "open_research_url": openResearch},
		},
		"trust_snapshot": map[string]any{
			"confidence_score": m.conf, "freshness_label": "Updated 5m ago",
			"last_human_review_label": "Apr 20", "disagreement_label": "Moderate",
		},
		"source_agreement": map[string]any{
			"independent_family_count": 8, "agreement_score_label": "High", "signal_breadth_label": "Broad",
			"open_research_sources_url": openResearch,
		},
		"credential_support": map[string]any{
			"strong_agent_support_label": "High", "top_clusters_label": "Long, Neutral",
			"speculative_share_label": "14%", "unclustered_share_label": "11%",
			"open_agent_discovery_url": openDiscover,
		},
		"methodology_stability": map[string]any{
			"weighting_model_status_label": "Stable", "last_formula_change_label": "21d ago",
			"sensitivity_label": "Low", "recompute_cadence_label": "15m", "dependency_risk_label": "Moderate",
			"open_methodology_url": openResearch, "open_research_url": openResearch,
		},
		"tabs": []any{"drivers", "topics", "cluster_breakdown", "research", "methodology"},
		"can_watchlist": true, "watchlist_locked": watchlistLocked,
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// floorIndexDetailMergeDBAndComposed overlays DB index entry onto the demo detail template when available.
func floorIndexDetailMergeDBAndComposed(e *dbpkg.FloorIndexEntry, watchlistLocked bool) map[string]any {
	base := floorComposedIndexDetailPageOnly(e.IndexID, watchlistLocked)
	if base == nil {
		base = map[string]any{
			"index_id": e.IndexID,
			"header": map[string]any{
				"title": e.Title, "subtitle": e.Subtitle, "type_label": e.Type,
				"access_tier": e.AccessTier, "timeframe": "7d",
				"can_watchlist": e.CanWatchlist, "watchlist_locked": watchlistLocked,
			},
			"hero": map[string]any{
				"thesis": e.WhyItMatters, "current_reading": e.CurrentReading,
				"open_floor_url": "/", "open_research_url": "/research",
			},
			"macro_strip":            floorMacroStripDefault(),
			"topic_contribution_rows": []any{},
			"what_moved":             []any{},
			"counter_evidence":       map[string]any{"severity_label": "—", "items": []any{}},
			"signals_to_watch":       []any{},
			"trust_snapshot":         floorDecodeJSONObject(e.TrustSnapshotJSON),
			"source_agreement": map[string]any{
				"open_research_sources_url": "/research",
			},
			"credential_support": map[string]any{
				"open_agent_discovery_url": "/discover?indexId=" + e.IndexID,
			},
			"methodology_stability": map[string]any{
				"open_methodology_url": "/research", "open_research_url": "/research",
			},
			"tabs": []any{"drivers", "topics", "cluster_breakdown", "research", "methodology"},
			"can_watchlist": e.CanWatchlist, "watchlist_locked": watchlistLocked,
		}
		return base
	}
	if h, ok := base["header"].(map[string]any); ok {
		h["title"] = e.Title
		if strings.TrimSpace(e.Subtitle) != "" {
			h["subtitle"] = e.Subtitle
		}
		h["type_label"] = e.Type
		h["access_tier"] = e.AccessTier
		h["can_watchlist"] = e.CanWatchlist
		h["watchlist_locked"] = watchlistLocked
	}
	if hero, ok := base["hero"].(map[string]any); ok {
		if strings.TrimSpace(e.WhyItMatters) != "" {
			hero["why_it_matters_now"] = e.WhyItMatters
		}
		if strings.TrimSpace(e.CurrentReading) != "" {
			hero["current_reading"] = e.CurrentReading
		}
	}
	if ts := floorDecodeJSONObject(e.TrustSnapshotJSON); len(ts) > 0 {
		base["trust_snapshot"] = ts
	}
	base["can_watchlist"] = e.CanWatchlist
	base["watchlist_locked"] = watchlistLocked
	return base
}

func (s *Server) handleFloorIndexDetail(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "indexID"))
	if id == "" {
		writeDetail(w, http.StatusBadRequest, "Missing index id")
		return
	}
	tier := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("tier")))
	locked := !(tier == "analytic" || tier == "terminal")
	dbq := s.dbCtx(r)
	var entry dbpkg.FloorIndexEntry
	err := dbq.Where("index_id = ?", id).First(&entry).Error
	if err == nil {
		writeJSON(w, http.StatusOK, floorIndexDetailMergeDBAndComposed(&entry, locked))
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		out := floorComposedIndexDetailPageOnly(id, locked)
		if out == nil {
			writeDetail(w, http.StatusNotFound, "Index not found")
			return
		}
		writeJSON(w, http.StatusOK, out)
		return
	}
	writeDetail(w, http.StatusInternalServerError, "DB error")
}
