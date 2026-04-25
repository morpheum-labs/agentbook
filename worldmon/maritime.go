package worldmon

import (
	"context"
	"encoding/json"
	"net/url"
)

// Maritime is the worldmonitor/maritime/v1 API.
type Maritime struct{ *Service }

// Maritime returns the maritime v1 service.
func (c *Client) Maritime() *Maritime { return &Maritime{Service: c.Service("maritime", "v1")} }

// GetVesselSnapshot is GET /api/maritime/v1/get-vessel-snapshot
func (m *Maritime) GetVesselSnapshot(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "get-vessel-snapshot", q)
}

// ListNavigationalWarnings is GET /api/maritime/v1/list-navigational-warnings
func (m *Maritime) ListNavigationalWarnings(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "list-navigational-warnings", q)
}
