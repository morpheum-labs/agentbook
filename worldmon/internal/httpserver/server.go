// Package httpserver is the local HTTP service for the worldmon client.
package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/morpheumlabs/agentbook/worldmon"
	"github.com/morpheumlabs/agentbook/worldmon/internal/regclient"
)

const (
	svcName        = "worldmon"
	svcDescription = "HTTP proxy for World Monitor (worldmonitor.app) service APIs; generic GET by service/version/method."
)

// Config holds environment-driven server settings.
type Config struct {
	Port              int
	WorldMonitorKey   string
	WorldMonitorBase  string
	PublicBaseURL     string
	RegistryBaseURL   string
	RegistryToken     string
	HeartbeatInterval time.Duration
	Version           string
}

// LoadConfig reads env: PORT, WORLDMONITOR_API_KEY, WORLDMONITOR_API_BASE, PUBLIC_BASE_URL, AGENTGLOBE_BASE_URL, SERVICE_REGISTRY_TOKEN, HEARTBEAT_INTERVAL, WORLDSERVER_VERSION.
func LoadConfig() Config {
	p := 8080
	if s := os.Getenv("PORT"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			p = n
		}
	}
	hb := 5 * time.Minute
	if s := os.Getenv("HEARTBEAT_INTERVAL"); s != "" {
		if d, err := time.ParseDuration(s); err == nil && d > 0 {
			hb = d
		}
	}
	v := os.Getenv("WORLDSERVER_VERSION")
	if v == "" {
		v = "0.0.0-dev"
	}
	return Config{
		Port:              p,
		WorldMonitorKey:   strings.TrimSpace(os.Getenv(worldmon.DefaultWorldMonitorKeyEnv)),
		WorldMonitorBase:  strings.TrimSpace(os.Getenv(worldmon.DefaultWorldMonitorBaseEnv)),
		PublicBaseURL:     strings.TrimRight(strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL")), "/"),
		RegistryBaseURL:   strings.TrimRight(strings.TrimSpace(os.Getenv("AGENTGLOBE_BASE_URL")), "/"),
		RegistryToken:     os.Getenv("SERVICE_REGISTRY_TOKEN"),
		HeartbeatInterval: hb,
		Version:           v,
	}
}

// NewRouter wires routes. c is the World Monitor client (non-nil; build in RunContext).
func NewRouter(c *worldmon.Client, publicBase, ver string, rcli *regclient.Client) *chi.Mux {
	if ver == "" {
		ver = "0.0.0-dev"
	}
	if publicBase == "" {
		publicBase = "http://127.0.0.1:8080"
	}
	r := chi.NewRouter()
	r.Get("/health", handleHealth(c, ver, publicBase))
	r.Get("/info", handleInfo(c, ver, publicBase))
	r.Get("/openapi.json", handleOpenapi())
	r.Get("/capabilities", handleCapabilities())
	r.Post("/register", handleRegister(c, rcli, publicBase, ver))
	r.Get("/v1/wm/{service}/{version}/{method}", handleWMProxy(c))
	return r
}

func handleHealth(cl *worldmon.Client, version, publicBase string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r
		w.Header().Set("Content-Type", "application/json")
		keyed := cl != nil && cl.APIKey() != ""
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "ok",
			"service": svcName,
			"version": version,
			"ready":   true,
			"public_base_url":   publicBase,
			"world_key_configured": keyed,
		})
	}
}

func handleInfo(cl *worldmon.Client, version, publicBase string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r
		w.Header().Set("Content-Type", "application/json")
		keyed := cl != nil && cl.APIKey() != ""
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name":        svcName,
			"version":     version,
			"description": svcDescription,
			"ready":       true,
			"public_base_url":   publicBase,
			"world_key_configured": keyed,
		})
	}
}

func handleOpenapi() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r
		b, err := readEmbeddedOpenapi()
		if err != nil || len(b) == 0 {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	}
}

func handleCapabilities() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r
		caps := make([]map[string]any, 0, len(ServiceNames))
		for _, d := range ServiceNames {
			caps = append(caps, map[string]any{
				"service":  d,
				"summary":  "GET /v1/wm/{service}/v1/{method} (version may be v1 or v2, etc.)",
				"path_pattern": "/v1/wm/" + d + "/v1/…",
			})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"service":      svcName,
			"domains":      ServiceNames,
			"count":        len(ServiceNames),
			"capabilities": caps,
		})
	}
}

func handleRegister(cl *worldmon.Client, rcli *regclient.Client, publicBase, version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r
		if rcli == nil || !rcli.Capable() {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":     false,
				"reason": "AGENTGLOBE_BASE_URL and SERVICE_REGISTRY_TOKEN not both set",
			})
			return
		}
		keyed := cl != nil && cl.APIKey() != ""
		req := regclient.RegisterRequest{
			Name:        svcName,
			Version:     version,
			BaseURL:     publicBase,
			Description: svcDescription,
			Category:    "world_monitor",
			Tags:        []string{"worldmonitor", "geopolitics", "news", "conflict", "intelligence"},
			Domains:     ServiceNames,
			OpenapiURL:  publicBase + "/openapi.json",
			Metadata: map[string]any{
				"kind": "worldmon_server", "upstream": "https://worldmonitor.app", "key_configured": keyed,
			},
		}
		if err := rcli.Register(r.Context(), req); err != nil {
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(map[string]string{"detail": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}
}

func handleWMProxy(cl *worldmon.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := strings.TrimSpace(chi.URLParam(r, "service"))
		ver := strings.TrimSpace(chi.URLParam(r, "version"))
		method := strings.TrimSpace(chi.URLParam(r, "method"))
		if svc == "" || ver == "" || method == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"detail": "service, version, and method are required in path"})
			return
		}
		_ = r.ParseForm()
		q := r.Form
		if q == nil {
			q = url.Values{}
		}
		b, err := cl.Service(svc, ver).Fetch(r.Context(), method, q)
		if err != nil {
			if cl.APIKey() == "" {
				w.WriteHeader(http.StatusUnauthorized)
			} else {
				w.WriteHeader(http.StatusBadGateway)
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"detail": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	}
}

// RunContext runs the HTTP server and background registry tasks.
// Caller must provide a [worldmon.Client] (e.g. from [worldmon.New] and optional [worldmon.WithBaseURL]).
func RunContext(ctx context.Context, cfg Config, c *worldmon.Client, rcli *regclient.Client, out io.Writer) error {
	if c == nil {
		return errors.New("httpserver: nil World Monitor client")
	}
	if out == nil {
		out = os.Stderr
	}
	pub := cfg.PublicBaseURL
	if pub == "" {
		pub = fmt.Sprintf("http://127.0.0.1:%d", cfg.Port)
	}
	h := NewRouter(c, pub, cfg.Version, rcli)
	if rcli != nil && rcli.Capable() {
		_, _ = io.WriteString(out, "httpserver: sending initial registry POST\n")
		keyed := c != nil && c.APIKey() != ""
		go func() {
			req := regclient.RegisterRequest{
				Name:        svcName,
				Version:     cfg.Version,
				BaseURL:     pub,
				Description: svcDescription,
				Category:    "world_monitor",
				Tags:        []string{"worldmonitor", "geopolitics"},
				Domains:     ServiceNames,
				OpenapiURL:  pub + "/openapi.json",
				Metadata: map[string]any{
					"kind": "worldmon_server", "upstream": c.BaseURL(), "key_configured": keyed,
				},
			}
			if err := rcli.Register(context.Background(), req); err != nil {
				logf(out, "httpserver: registry register failed: %v", err)
			} else {
				logf(out, "httpserver: registered with agentglobe")
			}
		}()
		tk := time.NewTicker(cfg.HeartbeatInterval)
		go func() {
			for {
				select {
				case <-ctx.Done():
					tk.Stop()
					return
				case <-tk.C:
					_ = rcli.Heartbeat(context.Background(), svcName, pub)
				}
			}
		}()
	}
	if rcli == nil {
		_, _ = io.WriteString(out, "httpserver: AGENTGLOBE_BASE_URL / SERVICE_REGISTRY_TOKEN not set; skipping registry\n")
	} else if !rcli.Capable() {
		_, _ = io.WriteString(out, "httpserver: registry token or base empty; skipping registry\n")
	}
	if c != nil && c.APIKey() == "" {
		_, _ = io.WriteString(out, "httpserver: "+worldmon.DefaultWorldMonitorKeyEnv+" is empty: upstream may return 401\n")
	}
	srv := &http.Server{Addr: fmt.Sprintf("0.0.0.0:%d", cfg.Port), Handler: h}
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()
	logf(out, "httpserver: listening on %s", srv.Addr)
	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func logf(w io.Writer, f string, a ...any) {
	s := time.Now().UTC().Format(time.RFC3339) + " " + fmt.Sprintf(f, a...) + "\n"
	_, _ = w.Write([]byte(s))
}
