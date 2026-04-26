package worldmon

import (
	"context"
	"encoding/json"
	"net/url"
)

// Military is GET /api/military/v1/…
type Military struct{ *Service }

// Military returns the military v1 service.
func (c *Client) Military() *Military { return &Military{Service: c.Service("military", "v1")} }

// ListMilitaryFlights is GET /api/military/v1/list-military-flights
func (m *Military) ListMilitaryFlights(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "list-military-flights", q)
}

// GetTheaterPosture is GET /api/military/v1/get-theater-posture
func (m *Military) GetTheaterPosture(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "get-theater-posture", q)
}

// GetAircraftDetails is GET /api/military/v1/get-aircraft-details
func (m *Military) GetAircraftDetails(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "get-aircraft-details", q)
}

// GetAircraftDetailsBatch is GET /api/military/v1/get-aircraft-details-batch
func (m *Military) GetAircraftDetailsBatch(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "get-aircraft-details-batch", q)
}

// GetWingbitsStatus is GET /api/military/v1/get-wingbits-status
func (m *Military) GetWingbitsStatus(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "get-wingbits-status", q)
}

// GetUSNIFleetReport is GET /api/military/v1/get-usni-fleet-report
func (m *Military) GetUSNIFleetReport(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "get-usni-fleet-report", q)
}

// ListMilitaryBases is GET /api/military/v1/list-military-bases
func (m *Military) ListMilitaryBases(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "list-military-bases", q)
}

// GetWingbitsLiveFlight is GET /api/military/v1/get-wingbits-live-flight
func (m *Military) GetWingbitsLiveFlight(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "get-wingbits-live-flight", q)
}

// ListDefensePatents is GET /api/military/v1/list-defense-patents
func (m *Military) ListDefensePatents(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "list-defense-patents", q)
}
