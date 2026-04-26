package worldmon

import (
	"context"
	"encoding/json"
	"net/url"
)

// Intelligence is GET /api/intelligence/v1/…
type Intelligence struct{ *Service }

// Intelligence returns the intelligence v1 service.
func (c *Client) Intelligence() *Intelligence {
	return &Intelligence{Service: c.Service("intelligence", "v1")}
}

// GetRiskScores is GET /api/intelligence/v1/get-risk-scores
func (i *Intelligence) GetRiskScores(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "get-risk-scores", q)
}

// GetCountryRisk is GET /api/intelligence/v1/get-country-risk
func (i *Intelligence) GetCountryRisk(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "get-country-risk", q)
}

// GetPizzintStatus is GET /api/intelligence/v1/get-pizzint-status
func (i *Intelligence) GetPizzintStatus(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "get-pizzint-status", q)
}

// ClassifyEvent is GET /api/intelligence/v1/classify-event
func (i *Intelligence) ClassifyEvent(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "classify-event", q)
}

// GetCountryIntelBrief is GET /api/intelligence/v1/get-country-intel-brief
func (i *Intelligence) GetCountryIntelBrief(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "get-country-intel-brief", q)
}

// SearchGdeltDocuments is GET /api/intelligence/v1/search-gdelt-documents
func (i *Intelligence) SearchGdeltDocuments(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "search-gdelt-documents", q)
}

// DeductSituation is GET /api/intelligence/v1/deduct-situation
func (i *Intelligence) DeductSituation(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "deduct-situation", q)
}

// GetCountryFacts is GET /api/intelligence/v1/get-country-facts
func (i *Intelligence) GetCountryFacts(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "get-country-facts", q)
}

// ListSecurityAdvisories is GET /api/intelligence/v1/list-security-advisories
func (i *Intelligence) ListSecurityAdvisories(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "list-security-advisories", q)
}

// ListSatellites is GET /api/intelligence/v1/list-satellites
func (i *Intelligence) ListSatellites(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "list-satellites", q)
}

// ListGpsInterference is GET /api/intelligence/v1/list-gps-interference
func (i *Intelligence) ListGpsInterference(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "list-gps-interference", q)
}

// ListOrefAlerts is GET /api/intelligence/v1/list-oref-alerts
func (i *Intelligence) ListOrefAlerts(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "list-oref-alerts", q)
}

// ListTelegramFeed is GET /api/intelligence/v1/list-telegram-feed
func (i *Intelligence) ListTelegramFeed(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "list-telegram-feed", q)
}

// GetCompanyEnrichment is GET /api/intelligence/v1/get-company-enrichment
func (i *Intelligence) GetCompanyEnrichment(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "get-company-enrichment", q)
}

// ListCompanySignals is GET /api/intelligence/v1/list-company-signals
func (i *Intelligence) ListCompanySignals(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "list-company-signals", q)
}

// GetGdeltTopicTimeline is GET /api/intelligence/v1/get-gdelt-topic-timeline
func (i *Intelligence) GetGdeltTopicTimeline(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "get-gdelt-topic-timeline", q)
}

// ListCrossSourceSignals is GET /api/intelligence/v1/list-cross-source-signals
func (i *Intelligence) ListCrossSourceSignals(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "list-cross-source-signals", q)
}

// ListMarketImplications is GET /api/intelligence/v1/list-market-implications
func (i *Intelligence) ListMarketImplications(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "list-market-implications", q)
}

// GetSocialVelocity is GET /api/intelligence/v1/get-social-velocity
func (i *Intelligence) GetSocialVelocity(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "get-social-velocity", q)
}

// GetCountryEnergyProfile is GET /api/intelligence/v1/get-country-energy-profile
func (i *Intelligence) GetCountryEnergyProfile(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "get-country-energy-profile", q)
}

// ComputeEnergyShockScenario is GET /api/intelligence/v1/compute-energy-shock
// (file compute-energy-shock.ts — path segment without the “Scenario” suffix).
func (i *Intelligence) ComputeEnergyShockScenario(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "compute-energy-shock", q)
}

// GetCountryPortActivity is GET /api/intelligence/v1/get-country-port-activity
func (i *Intelligence) GetCountryPortActivity(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "get-country-port-activity", q)
}

// GetRegionalSnapshot is GET /api/intelligence/v1/get-regional-snapshot
func (i *Intelligence) GetRegionalSnapshot(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "get-regional-snapshot", q)
}

// GetRegimeHistory is GET /api/intelligence/v1/get-regime-history
func (i *Intelligence) GetRegimeHistory(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "get-regime-history", q)
}

// GetRegionalBrief is GET /api/intelligence/v1/get-regional-brief
func (i *Intelligence) GetRegionalBrief(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return i.Fetch(ctx, "get-regional-brief", q)
}
