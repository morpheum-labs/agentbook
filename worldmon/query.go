package worldmon

import "net/url"

// RiskScoresByRegion builds a common "region" query (e.g. for get-risk-scores).
func RiskScoresByRegion(region string) url.Values {
	v := url.Values{}
	if region != "" {
		v.Set("region", region)
	}
	return v
}
