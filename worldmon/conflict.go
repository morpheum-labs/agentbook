package worldmon

import (
	"context"
	"encoding/json"
	"net/url"
)

// Conflict is GET /api/conflict/v1/…
type Conflict struct{ *Service }

// Conflict returns the conflict v1 service.
func (c *Client) Conflict() *Conflict { return &Conflict{Service: c.Service("conflict", "v1")} }

// ListAcledEvents is GET /api/conflict/v1/list-acled-events
func (c *Conflict) ListAcledEvents(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return c.Fetch(ctx, "list-acled-events", q)
}

// ListUcdpEvents is GET /api/conflict/v1/list-ucdp-events
func (c *Conflict) ListUcdpEvents(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return c.Fetch(ctx, "list-ucdp-events", q)
}

// GetHumanitarianSummary is GET /api/conflict/v1/get-humanitarian-summary
func (c *Conflict) GetHumanitarianSummary(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return c.Fetch(ctx, "get-humanitarian-summary", q)
}

// GetHumanitarianSummaryBatch is GET /api/conflict/v1/get-humanitarian-summary-batch
func (c *Conflict) GetHumanitarianSummaryBatch(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return c.Fetch(ctx, "get-humanitarian-summary-batch", q)
}

// ListIranEvents is GET /api/conflict/v1/list-iran-events
func (c *Conflict) ListIranEvents(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return c.Fetch(ctx, "list-iran-events", q)
}
