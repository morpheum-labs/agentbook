package worldmon

import (
	"context"
	"encoding/json"
	"net/url"
)

// ShippingV2 is the worldmonitor/shipping/v2 API (this tree only has v2 today on GitHub).
type ShippingV2 struct{ *Service }

// ShippingV2 returns the shipping v2 service.
func (c *Client) ShippingV2() *ShippingV2 { return &ShippingV2{Service: c.Service("shipping", "v2")} }

// RouteIntelligence is GET /api/shipping/v2/route-intelligence
func (s *ShippingV2) RouteIntelligence(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return s.Fetch(ctx, "route-intelligence", q)
}

// RegisterWebhook is GET /api/shipping/v2/register-webhook
func (s *ShippingV2) RegisterWebhook(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return s.Fetch(ctx, "register-webhook", q)
}

// ListWebhooks is GET /api/shipping/v2/list-webhooks
func (s *ShippingV2) ListWebhooks(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return s.Fetch(ctx, "list-webhooks", q)
}
