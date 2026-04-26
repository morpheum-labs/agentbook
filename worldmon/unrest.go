package worldmon

import (
	"context"
	"encoding/json"
	"net/url"
)

// Unrest is GET /api/unrest/v1/…
type Unrest struct{ *Service }

// Unrest returns the unrest v1 service.
func (c *Client) Unrest() *Unrest { return &Unrest{Service: c.Service("unrest", "v1")} }

// ListUnrestEvents is GET /api/unrest/v1/list-unrest-events
func (u *Unrest) ListUnrestEvents(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return u.Fetch(ctx, "list-unrest-events", q)
}
