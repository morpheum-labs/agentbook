package worldmon

import (
	"context"
	"encoding/json"
	"net/url"
)

// Climate is the worldmonitor/climate/v1 API.
type Climate struct{ *Service }

// Climate returns the climate v1 service.
func (c *Client) Climate() *Climate { return &Climate{Service: c.Service("climate", "v1")} }

// GetCo2Monitoring is GET /api/climate/v1/get-co2-monitoring
func (c *Climate) GetCo2Monitoring(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return c.Fetch(ctx, "get-co2-monitoring", q)
}

// GetOceanIceData is GET /api/climate/v1/get-ocean-ice-data
func (c *Climate) GetOceanIceData(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return c.Fetch(ctx, "get-ocean-ice-data", q)
}

// ListAirQualityData is GET /api/climate/v1/list-air-quality-data
func (c *Climate) ListAirQualityData(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return c.Fetch(ctx, "list-air-quality-data", q)
}

// ListClimateAnomalies is GET /api/climate/v1/list-climate-anomalies
func (c *Climate) ListClimateAnomalies(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return c.Fetch(ctx, "list-climate-anomalies", q)
}

// ListClimateDisasters is GET /api/climate/v1/list-climate-disasters
func (c *Climate) ListClimateDisasters(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return c.Fetch(ctx, "list-climate-disasters", q)
}

// ListClimateNews is GET /api/climate/v1/list-climate-news
func (c *Climate) ListClimateNews(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return c.Fetch(ctx, "list-climate-news", q)
}
