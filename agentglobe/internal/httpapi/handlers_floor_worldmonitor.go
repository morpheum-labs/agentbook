package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/worldmonitor"
	"gorm.io/gorm"
)

const floorWorldMonitorCacheTTL = 15 * time.Minute

func floorWMRegionForQuestion(q *dbpkg.FloorQuestion) string {
	if q.WmContextID != nil && strings.TrimSpace(*q.WmContextID) != "" {
		return strings.TrimSpace(*q.WmContextID)
	}
	parts := strings.Split(q.Category, "/")
	if len(parts) > 1 {
		return strings.TrimSpace(parts[len(parts)-1])
	}
	return strings.TrimSpace(q.Category)
}

func floorWMContextFromSignal(sig *dbpkg.FloorExternalSignal) map[string]any {
	return map[string]any{
		"instability": floorDecodeJSONObject(sig.InstabilityIndexJSON),
		"convergence": floorDecodeJSONObject(sig.GeoConvergenceJSON),
		"forecast":    floorDecodeJSONObject(sig.ForecastSummaryJSON),
	}
}

func floorWMAlertsFromConvergenceJSON(geoJSON string) []map[string]any {
	conv := floorDecodeJSONObject(geoJSON)
	out := make([]map[string]any, 0, len(conv))
	for region, raw := range conv {
		vm, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		gc, _ := vm["geo_convergence"].(float64)
		if gc >= 70 {
			out = append(out, map[string]any{
				"type":    "wm_high_geo_convergence",
				"region":  region,
				"score":   gc,
				"message": "WorldMonitor geographic convergence elevated — compare with AgentFloor cluster positions.",
			})
		}
	}
	return out
}

func floorLatestWorldMonitorSignal(db *gorm.DB, questionID string) (*dbpkg.FloorExternalSignal, error) {
	var s dbpkg.FloorExternalSignal
	err := db.Where("question_id = ? AND source = ?", questionID, "worldmonitor").
		Where("fetch_error IS NULL OR fetch_error = ?", "").
		Order("fetched_at DESC").Limit(1).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func floorExternalSignalCacheMeta(sig *dbpkg.FloorExternalSignal) map[string]any {
	m := map[string]any{
		"signal_id":  sig.ID,
		"fetched_at": sig.FetchedAt.UTC().Format(time.RFC3339Nano),
		"source":     sig.Source,
	}
	if sig.UpstreamSignatureMs != nil {
		m["upstream_signature_ms"] = *sig.UpstreamSignatureMs
	} else {
		m["upstream_signature_ms"] = nil
	}
	if sig.FetchError != nil && strings.TrimSpace(*sig.FetchError) != "" {
		m["fetch_error"] = *sig.FetchError
	} else {
		m["fetch_error"] = nil
	}
	return m
}

// GET /api/v1/floor/questions/{questionID}/context/worldmonitor
// Terminal tier (v1 stub): any authenticated agent until entitlements are enforced.
func (s *Server) handleFloorQuestionWorldMonitorContext(w http.ResponseWriter, r *http.Request) {
	a := s.requireAgent(w, r)
	if a == nil {
		return
	}
	_ = a
	qid := strings.TrimSpace(chi.URLParam(r, "questionID"))
	if qid == "" {
		writeDetail(w, http.StatusBadRequest, "Missing question id")
		return
	}
	db := s.dbCtx(r)
	var q dbpkg.FloorQuestion
	if err := db.First(&q, "id = ?", qid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeDetail(w, http.StatusNotFound, "Question not found")
			return
		}
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	refresh := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("refresh")), "true") ||
		strings.TrimSpace(r.URL.Query().Get("refresh")) == "1"
	region := floorWMRegionForQuestion(&q)

	if !refresh {
		if prev, err := floorLatestWorldMonitorSignal(db, qid); err == nil && prev != nil {
			if time.Since(prev.FetchedAt) < floorWorldMonitorCacheTTL {
				writeJSON(w, http.StatusOK, floorWorldMonitorResponse(&q, region, prev, true, false))
				return
			}
		} else if err != nil {
			writeDetail(w, http.StatusInternalServerError, "DB error")
			return
		}
	}

	key := worldmonitor.APIKey()
	if key == "" {
		writeJSON(w, http.StatusOK, map[string]any{
			"question_id":      q.ID,
			"wm_context_id":    q.WmContextID,
			"wm_query_region":  region,
			"terminal_tier":    "stub_any_authenticated_agent",
			"live":             false,
			"reason":           "WORLDMONITOR_API_KEY not configured on server",
			"context":          map[string]any{"worldmonitor": nil},
			"cache":            nil,
			"alerts":           []map[string]any{},
			"authenticated_as": a.ID,
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 14*time.Second)
	defer cancel()
	wm := worldmonitor.NewClient()
	riskB, errR := wm.FetchRiskScores(ctx, region)
	fcB, errF := wm.FetchForecasts(ctx, "", region)
	if errR != nil && errF != nil {
		if prev, perr := floorLatestWorldMonitorSignal(db, qid); perr == nil && prev != nil {
			writeJSON(w, http.StatusOK, floorWorldMonitorResponse(&q, region, prev, true, true))
			return
		}
		writeDetail(w, http.StatusBadGateway, "WorldMonitor upstream error: "+errR.Error())
		return
	}
	bundle, normErr := worldmonitor.NormalizeBundle(riskB, fcB)
	if normErr != nil {
		writeDetail(w, http.StatusBadGateway, "WorldMonitor normalize error")
		return
	}
	rawObj := map[string]any{}
	if len(riskB) > 0 {
		var v any
		_ = json.Unmarshal(riskB, &v)
		rawObj["risk_scores"] = v
	}
	if len(fcB) > 0 {
		var v any
		_ = json.Unmarshal(fcB, &v)
		rawObj["forecasts"] = v
	}
	rawBytes, _ := json.Marshal(rawObj)
	instBytes, _ := json.Marshal(bundle.Instability)
	geoBytes, _ := json.Marshal(bundle.Convergence)
	fcOutBytes, _ := json.Marshal(bundle.Forecast)
	var upMs *int64
	if bundle.UpstreamSigMs > 0 {
		ms := bundle.UpstreamSigMs
		upMs = &ms
	}
	topic := q.Category
	sig := dbpkg.FloorExternalSignal{
		ID:                   uuid.NewString(),
		QuestionID:           &q.ID,
		TopicClass:           &topic,
		FetchedAt:            time.Now().UTC(),
		Source:               "worldmonitor",
		RawDataJSON:          string(rawBytes),
		InstabilityIndexJSON: string(instBytes),
		GeoConvergenceJSON:   string(geoBytes),
		ForecastSummaryJSON:  string(fcOutBytes),
		UpstreamSignatureMs:  upMs,
	}
	if err := db.Create(&sig).Error; err != nil {
		writeDetail(w, http.StatusInternalServerError, "DB error")
		return
	}
	writeJSON(w, http.StatusOK, floorWorldMonitorResponse(&q, region, &sig, false, false))
}

func floorExternalSignalsDigestBlock(sig *dbpkg.FloorExternalSignal) map[string]any {
	return map[string]any{
		"signal_id":  sig.ID,
		"fetched_at": sig.FetchedAt.UTC().Format(time.RFC3339Nano),
		"source":     sig.Source,
		"context":    map[string]any{"worldmonitor": floorWMContextFromSignal(sig)},
	}
}

func floorDigestAttachExternalSignals(db *gorm.DB, row map[string]any) {
	qid, _ := row["question_id"].(string)
	if qid == "" {
		row["external_signals"] = []any{}
		return
	}
	sig, err := floorLatestWorldMonitorSignal(db, qid)
	if err != nil || sig == nil {
		row["external_signals"] = []any{}
		return
	}
	row["external_signals"] = []any{floorExternalSignalsDigestBlock(sig)}
}

func floorWorldMonitorResponse(q *dbpkg.FloorQuestion, region string, sig *dbpkg.FloorExternalSignal, cacheHit, stale bool) map[string]any {
	wmCtx := floorWMContextFromSignal(sig)
	live := sig.FetchError == nil || (sig.FetchError != nil && strings.TrimSpace(*sig.FetchError) == "")
	resp := map[string]any{
		"question_id":     q.ID,
		"wm_query_region": region,
		"terminal_tier":   "stub_any_authenticated_agent",
		"cache_hit":       cacheHit,
		"stale_upstream":  stale,
		"cache":           floorExternalSignalCacheMeta(sig),
		"context":         map[string]any{"worldmonitor": wmCtx},
		"alerts":          floorWMAlertsFromConvergenceJSON(sig.GeoConvergenceJSON),
		"wm_context_id":   nil,
		"live":            live,
	}
	if q.WmContextID != nil {
		resp["wm_context_id"] = *q.WmContextID
	} else {
		resp["wm_context_id"] = nil
	}
	return resp
}
