package worldmon

import "net/url"

// RiskScoresByRegion is a small helper for the common "region" query on get-risk-scores.
func RiskScoresByRegion(region string) url.Values {
	v := url.Values{}
	if region != "" {
		v.Set("region", region)
	}
	return v
}

// ForecastsByRegionDomain is a small helper for get-forecasts style queries.
func ForecastsByRegionDomain(domain, region string) url.Values {
	v := url.Values{}
	if domain != "" {
		v.Set("domain", domain)
	}
	if region != "" {
		v.Set("region", region)
	}
	return v
}
