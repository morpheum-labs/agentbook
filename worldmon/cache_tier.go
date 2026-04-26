package worldmon

import (
	"strings"
)

// CacheTier is a client-side label for per-route cache/refresh heuristics (e.g. fast
// vs slow revalidation). Paths should match a GET to /api/{service}/{version}/{method}.
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
