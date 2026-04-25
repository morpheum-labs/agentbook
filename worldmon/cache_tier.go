package worldmon

import (
	"strings"
)

// CacheTier matches the public edge gateway’s RPC cache tier in
// [server/gateway.ts] (TIER_HEADERS / RPC_CACHE_TIER). Use for client-side
// heuristics; production responses may add premium overrides (see PREMIUM_RPC_PATHS
// in the same repo). Paths should match a GET to /api/{service}/{version}/{method}.
//
// [server/gateway.ts]: https://github.com/koala73/worldmonitor/blob/main/server/gateway.ts
type CacheTier string

const (
	CacheTierFast        CacheTier = "fast"
	CacheTierMedium      CacheTier = "medium"
	CacheTierSlow        CacheTier = "slow"
	CacheTierSlowBrowser CacheTier = "slow-browser"
	CacheTierStatic      CacheTier = "static"
	CacheTierDaily       CacheTier = "daily"
	CacheTierNoStore     CacheTier = "no-store"
)

func (t CacheTier) String() string { return string(t) }

// APIPath is the request path the [Client] uses for GET
// /api/{service}/{version}/{method} (version defaults to "v1" if empty).
func APIPath(service, version, method string) string {
	s := strings.Trim(strings.TrimSpace(service), "/")
	m := strings.Trim(strings.TrimSpace(method), "/")
	v := strings.TrimSpace(version)
	if v == "" {
		v = "v1"
	}
	if s == "" || m == "" {
		return ""
	}
	return "/api/" + s + "/" + v + "/" + m
}

// cacheTierPathKey normalizes the path the same way the edge gateway does for
// the RPC map (trailing slash stripped except for root).
func cacheTierPathKey(p string) string {
	p = strings.TrimSpace(p)
	if len(p) > 1 && strings.HasSuffix(p, "/") {
		return p[:len(p)-1]
	}
	return p
}

// CacheTierForPath returns the gateway’s declared RPC cache tier for a full
// path, if present. Unknown paths return false. Both /api/…/v1/… and, when
// listed, legacy /api/v2/shipping/… forms are in [rpcCacheTier].
func CacheTierForPath(path string) (CacheTier, bool) {
	key := cacheTierPathKey(path)
	if key == "" {
		return "", false
	}
	t, ok := rpcCacheTier[key]
	return t, ok
}

// MethodCacheTier is like [CacheTierForPath] for this service’s
// [Service.Fetch] method (last path segment, kebab-case file stem).
func (s *Service) MethodCacheTier(method string) (CacheTier, bool) {
	if s == nil || s.client == nil {
		return "", false
	}
	if strings.TrimSpace(s.name) == "" {
		return "", false
	}
	p := APIPath(s.name, s.version, method)
	if p == "" {
		return "", false
	}
	return CacheTierForPath(p)
}
