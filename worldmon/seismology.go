package worldmon

import (
	"context"
	"encoding/json"
	"net/url"
)

// Seismology is GET /api/seismology/v1/…
type Seismology struct{ *Service }

// Seismology returns the seismology v1 service.
func (c *Client) Seismology() *Seismology { return &Seismology{Service: c.Service("seismology", "v1")} }

// ListEarthquakes is GET /api/seismology/v1/list-earthquakes
func (s *Seismology) ListEarthquakes(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return s.Fetch(ctx, "list-earthquakes", q)
}
