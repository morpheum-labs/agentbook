package worldmonitor

import (
	"encoding/json"
	"math"
	"strings"
)

// Bundle is the normalized AgentFloor-facing slice of WorldMonitor data (F7-friendly JSON shapes).
type Bundle struct {
	Instability   map[string]int                `json:"instability"`
	Convergence   map[string]map[string]float64 `json:"convergence"`
	Forecast      map[string]any                `json:"forecast"`
	RawEnvelope   map[string]json.RawMessage    `json:"-"` // not serialized in API; stored in DB raw_data
	UpstreamSigMs int64                         `json:"-"` // max computedAt / generatedAt from payloads
}

type riskScoresPayload struct {
	CiiScores      []ciiScoreJSON      `json:"ciiScores"`
	StrategicRisks []strategicRiskJSON `json:"strategicRisks"`
}

type ciiScoreJSON struct {
	Region        string       `json:"region"`
	CombinedScore float64      `json:"combinedScore"`
	Components    *ciiCompJSON `json:"components"`
	ComputedAt    int64        `json:"computedAt"`
}

type ciiCompJSON struct {
	GeoConvergence float64 `json:"geoConvergence"`
}

type strategicRiskJSON struct {
	Region string  `json:"region"`
	Score  float64 `json:"score"`
}

type forecastsPayload struct {
	Forecasts   []wmForecastRow `json:"forecasts"`
	GeneratedAt int64           `json:"generatedAt"`
}

type wmForecastRow struct {
	ID          string  `json:"id"`
	Region      string  `json:"region"`
	Title       string  `json:"title"`
	Probability float64 `json:"probability"`
	TimeHorizon string  `json:"timeHorizon"`
	Scenario    string  `json:"scenario"`
}

// NormalizeBundle parses WorldMonitor JSON bodies into the digest/UI-friendly context.worldmonitor shape.
func NormalizeBundle(riskJSON, forecastJSON []byte) (*Bundle, error) {
	b := &Bundle{
		Instability: make(map[string]int),
		Convergence: make(map[string]map[string]float64),
		Forecast:    map[string]any{},
		RawEnvelope: make(map[string]json.RawMessage),
	}
	if len(riskJSON) > 0 {
		b.RawEnvelope["risk_scores"] = json.RawMessage(riskJSON)
		var rs riskScoresPayload
		if err := json.Unmarshal(riskJSON, &rs); err == nil {
			for _, c := range rs.CiiScores {
				k := strings.TrimSpace(c.Region)
				if k == "" {
					continue
				}
				b.Instability[k] = int(math.Round(c.CombinedScore))
				conv := map[string]float64{"combined": math.Round(c.CombinedScore*10) / 10}
				if c.Components != nil {
					conv["geo_convergence"] = math.Round(c.Components.GeoConvergence*10) / 10
				}
				b.Convergence[k] = conv
				if c.ComputedAt > b.UpstreamSigMs {
					b.UpstreamSigMs = c.ComputedAt
				}
			}
			for _, s := range rs.StrategicRisks {
				k := strings.TrimSpace(s.Region)
				if k == "" {
					continue
				}
				if _, ok := b.Convergence[k]; !ok {
					b.Convergence[k] = map[string]float64{}
				}
				b.Convergence[k]["strategic_risk"] = math.Round(s.Score*10) / 10
			}
		}
	}
	if len(forecastJSON) > 0 {
		b.RawEnvelope["forecasts"] = json.RawMessage(forecastJSON)
		var fs forecastsPayload
		if err := json.Unmarshal(forecastJSON, &fs); err == nil {
			if fs.GeneratedAt > b.UpstreamSigMs {
				b.UpstreamSigMs = fs.GeneratedAt
			}
			var best *wmForecastRow
			for i := range fs.Forecasts {
				f := &fs.Forecasts[i]
				if best == nil || f.Probability > best.Probability {
					best = f
				}
			}
			if best != nil {
				b.Forecast["id"] = best.ID
				b.Forecast["title"] = best.Title
				b.Forecast["region"] = best.Region
				b.Forecast["probability"] = best.Probability
				b.Forecast["horizon"] = best.TimeHorizon
				b.Forecast["scenario"] = best.Scenario
			}
		}
	}
	return b, nil
}
