package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

const clawgotchaWebhookPathSuffix = "/webhook/clawgotcha"

// mergeRegisterIngressMetadata adds request-path / proxy hints under clawgotcha_register_ingress.
func mergeRegisterIngressMetadata(raw json.RawMessage, r *http.Request) json.RawMessage {
	var base map[string]any
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &base)
	}
	if base == nil {
		base = map[string]any{}
	}
	ing := map[string]any{
		"request_path": r.URL.Path,
	}
	if p := strings.TrimSpace(r.Header.Get("X-Forwarded-Prefix")); p != "" {
		ing["x_forwarded_prefix"] = p
	}
	if p := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); p != "" {
		ing["x_forwarded_proto"] = p
	}
	if h := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); h != "" {
		ing["x_forwarded_host"] = h
	}
	if h := strings.TrimSpace(r.Host); h != "" {
		ing["host"] = h
	}
	base["clawgotcha_register_ingress"] = ing
	out, err := json.Marshal(base)
	if err != nil {
		return raw
	}
	return out
}

// effectiveInstancePublicURL returns body public_url when set; otherwise derives gateway origin from callback_url.
func effectiveInstancePublicURL(publicURL *string, callbackURL string) *string {
	if publicURL != nil {
		if t := strings.TrimSpace(*publicURL); t != "" {
			return publicURL
		}
	}
	if derived, ok := publicURLFromClawgotchaCallback(callbackURL); ok {
		return &derived
	}
	return nil
}

// publicURLFromClawgotchaCallback strips ZeroClaw/MiroClaw clawgotcha webhook suffix from callback URL.
func publicURLFromClawgotchaCallback(callbackURL string) (string, bool) {
	raw := strings.TrimSpace(callbackURL)
	if raw == "" {
		return "", false
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", false
	}
	path := strings.TrimSuffix(u.Path, "/")
	if !strings.HasSuffix(path, clawgotchaWebhookPathSuffix) {
		return "", false
	}
	path = strings.TrimSuffix(path, clawgotchaWebhookPathSuffix)
	path = strings.TrimSuffix(path, "/")
	out := url.URL{Scheme: u.Scheme, Host: u.Host}
	if path != "" && path != "/" {
		out.Path = path
	}
	s := strings.TrimSuffix(out.String(), "/")
	return s, true
}
