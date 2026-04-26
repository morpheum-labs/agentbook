package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const defaultWMBaseURL = "https://worldmonitor.app"

func wmAPIBase() string {
	v := strings.TrimSpace(os.Getenv("WORLDMONITOR_API_BASE"))
	if v == "" {
		return defaultWMBaseURL
	}
	return strings.TrimRight(v, "/")
}

func wmAPIKey() string {
	return strings.TrimSpace(os.Getenv("WORLDMONITOR_API_KEY"))
}

type wmClient struct {
	baseURL string
	key     string
	http    *http.Client
}

func newWMClient() *wmClient {
	return &wmClient{
		baseURL: wmAPIBase(),
		key:     wmAPIKey(),
		http: &http.Client{
			Timeout: 12 * time.Second,
		},
	}
}

func (c *wmClient) getJSON(ctx context.Context, path string, query url.Values) (json.RawMessage, int, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, 0, err
	}
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, 0, err
	}
	if strings.TrimSpace(c.key) != "" {
		req.Header.Set("X-WorldMonitor-Key", c.key)
	}
	req.Header.Set("Accept", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return nil, res.StatusCode, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return b, res.StatusCode, fmt.Errorf("worldmonitor: HTTP %d", res.StatusCode)
	}
	return json.RawMessage(b), res.StatusCode, nil
}

func (c *wmClient) fetchRiskScores(ctx context.Context, region string) (json.RawMessage, error) {
	q := url.Values{}
	if strings.TrimSpace(region) != "" {
		q.Set("region", strings.TrimSpace(region))
	}
	body, code, err := c.getJSON(ctx, "/api/intelligence/v1/get-risk-scores", q)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("worldmonitor get-risk-scores: status %d body=%s", code, string(body))
	}
	return body, nil
}

func (c *wmClient) fetchForecasts(ctx context.Context, domain, region string) (json.RawMessage, error) {
	q := url.Values{}
	if strings.TrimSpace(domain) != "" {
		q.Set("domain", strings.TrimSpace(domain))
	}
	if strings.TrimSpace(region) != "" {
		q.Set("region", strings.TrimSpace(region))
	}
	body, code, err := c.getJSON(ctx, "/api/forecast/v1/get-forecasts", q)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("worldmonitor get-forecasts: status %d body=%s", code, string(body))
	}
	return body, nil
}

type wmBundle struct {
	Instability   map[string]int                `json:"instability"`
	Convergence   map[string]map[string]float64 `json:"convergence"`
	Forecast      map[string]any                `json:"forecast"`
	RawEnvelope   map[string]json.RawMessage    `json:"-"`
	UpstreamSigMs int64                         `json:"-"`
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

func wmNormalizeBundle(riskJSON, forecastJSON []byte) (*wmBundle, error) {
	b := &wmBundle{
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
