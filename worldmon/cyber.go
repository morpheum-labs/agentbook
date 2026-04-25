package worldmon

import (
	"context"
	"encoding/json"
	"net/url"
)

// Cyber is the worldmonitor/cyber/v1 API.
type Cyber struct{ *Service }

// Cyber returns the cyber v1 service.
func (c *Client) Cyber() *Cyber { return &Cyber{Service: c.Service("cyber", "v1")} }

// ListCyberThreats is GET /api/cyber/v1/list-cyber-threats
func (c *Cyber) ListCyberThreats(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return c.Fetch(ctx, "list-cyber-threats", q)
}
