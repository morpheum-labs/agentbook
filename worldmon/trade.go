package worldmon

import (
	"context"
	"encoding/json"
	"net/url"
)

// Trade is the worldmonitor/trade/v1 API.
type Trade struct{ *Service }

// Trade returns the trade v1 service.
func (c *Client) Trade() *Trade { return &Trade{Service: c.Service("trade", "v1")} }

// GetTradeRestrictions is GET /api/trade/v1/get-trade-restrictions
func (t *Trade) GetTradeRestrictions(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return t.Fetch(ctx, "get-trade-restrictions", q)
}

// GetTariffTrends is GET /api/trade/v1/get-tariff-trends
func (t *Trade) GetTariffTrends(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return t.Fetch(ctx, "get-tariff-trends", q)
}

// GetTradeFlows is GET /api/trade/v1/get-trade-flows
func (t *Trade) GetTradeFlows(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return t.Fetch(ctx, "get-trade-flows", q)
}

// GetTradeBarriers is GET /api/trade/v1/get-trade-barriers
func (t *Trade) GetTradeBarriers(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return t.Fetch(ctx, "get-trade-barriers", q)
}

// GetCustomsRevenue is GET /api/trade/v1/get-customs-revenue
func (t *Trade) GetCustomsRevenue(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return t.Fetch(ctx, "get-customs-revenue", q)
}

// ListComtradeFlows is GET /api/trade/v1/list-comtrade-flows
func (t *Trade) ListComtradeFlows(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return t.Fetch(ctx, "list-comtrade-flows", q)
}
