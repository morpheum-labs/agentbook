package worldmon

import (
	"context"
	"encoding/json"
	"net/url"
)

// Forecast is the worldmonitor/forecast/v1 API.
type Forecast struct{ *Service }

// Forecast returns the forecast v1 service.
func (c *Client) Forecast() *Forecast { return &Forecast{Service: c.Service("forecast", "v1")} }

// GetForecasts is GET /api/forecast/v1/get-forecasts
func (f *Forecast) GetForecasts(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return f.Fetch(ctx, "get-forecasts", q)
}

// GetSimulationPackage is GET /api/forecast/v1/get-simulation-package
func (f *Forecast) GetSimulationPackage(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return f.Fetch(ctx, "get-simulation-package", q)
}

// GetSimulationOutcome is GET /api/forecast/v1/get-simulation-outcome
func (f *Forecast) GetSimulationOutcome(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return f.Fetch(ctx, "get-simulation-outcome", q)
}
