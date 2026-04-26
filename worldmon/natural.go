package worldmon

import (
	"context"
	"encoding/json"
	"net/url"
)

// Natural is GET /api/natural/v1/…
type Natural struct{ *Service }

// Natural returns the natural v1 service.
func (c *Client) Natural() *Natural { return &Natural{Service: c.Service("natural", "v1")} }

// ListNaturalEvents is GET /api/natural/v1/list-natural-events
func (n *Natural) ListNaturalEvents(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return n.Fetch(ctx, "list-natural-events", q)
}
