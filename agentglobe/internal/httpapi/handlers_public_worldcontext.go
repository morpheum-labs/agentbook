package httpapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/morpheumlabs/agentbook/agentglobe/internal/worldcontext"
)

// GET /api/v1/public/world-context
// Public read proxy: forwards to the configured worldmon process at GET /v1/wm/{service}/{version}/{method}
// with query string built from remaining parameters. Applies rss_lib from config first; clients may override
// (same semantics as the former direct MCP path).
func (s *Server) handlePublicWorldContext(w http.ResponseWriter, r *http.Request) {
	in := r.URL.Query()
	method := firstQueryValueCI(in, "method")
	if method == "" {
		writeDetail(w, http.StatusBadRequest, "query parameter method is required (e.g. list-feed-digest)")
		return
	}
	base, err := worldcontext.ResolveServiceBase(s.dbCtx(r))
	if err != nil {
		writeDetail(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	svc := firstQueryValueCI(in, "service")
	if svc == "" {
		svc = "news"
	}
	ver := firstQueryValueCI(in, "version")
	if ver == "" {
		ver = "v1"
	}
	rel := fmt.Sprintf("/v1/wm/%s/%s/%s", url.PathEscape(svc), url.PathEscape(ver), strings.TrimLeft(method, "/"))
	u, err := url.Parse(base + rel)
	if err != nil {
		writeDetail(w, http.StatusInternalServerError, "bad upstream URL")
		return
	}
	merge := url.Values{}
	worldcontext.ApplyRssLibQuery(merge, s.Cfg.ResolvedRssLibPath(strings.TrimSpace(os.Getenv("CONFIG_PATH"))))
	for k, vals := range in {
		if queryKeyIsOneOfCI(k, "method", "service", "version") {
			continue
		}
		for _, v := range vals {
			// Do not set empty values: they would wipe defaults from ApplyRssLibQuery.
			if strings.TrimSpace(v) == "" {
				continue
			}
			merge.Set(k, v)
		}
	}
	u.RawQuery = merge.Encode()

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		writeDetail(w, http.StatusInternalServerError, err.Error())
		return
	}
	req.Header.Set("User-Agent", "agentglobe/1.0 (public world-context)")

	res, err := publicWorldContextHTTP().Do(req)
	if err != nil {
		writeDetail(w, http.StatusBadGateway, fmt.Sprintf("worldmon: %v", err))
		return
	}
	defer res.Body.Close()
	if ct := res.Header.Get("Content-Type"); strings.TrimSpace(ct) != "" {
		w.Header().Set("Content-Type", ct)
	} else {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(res.StatusCode)
	lr := io.LimitReader(res.Body, 32<<20)
	if _, err := io.Copy(w, lr); err != nil {
		// Response may be partial; still drain the socket for keep-alive.
		_, _ = io.Copy(io.Discard, res.Body)
		return
	}
	// If worldmon sent more than 32MiB, read the rest so the client can reuse the connection.
	_, _ = io.Copy(io.Discard, res.Body)
}

func publicWorldContextHTTP() *http.Client {
	// No client-level timeout: request context controls deadline (large digest responses).
	return &http.Client{}
}

func firstQueryValueCI(in url.Values, name string) string {
	ln := strings.ToLower(strings.TrimSpace(name))
	for k, vs := range in {
		if strings.ToLower(k) != ln {
			continue
		}
		for _, v := range vs {
			if s := strings.TrimSpace(v); s != "" {
				return s
			}
		}
	}
	return ""
}

func queryKeyIsOneOfCI(k string, names ...string) bool {
	lk := strings.ToLower(k)
	for _, n := range names {
		if lk == strings.ToLower(n) {
			return true
		}
	}
	return false
}
