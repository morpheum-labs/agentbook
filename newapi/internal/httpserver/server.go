// Package httpserver is the local HTTP service for the newapi client library.
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
	"github.com/morpheumlabs/agentbook/newapi"
	"github.com/morpheumlabs/agentbook/newapi/internal/regclient"
)

const (
	svcName        = "newapi"
	svcDescription = "HTTP service wrapping the News API (newsapi.org) v1 and v2 client for agents."
)

// Config holds environment-driven settings for the HTTP server.
type Config struct {
	Port              int
	NewsAPIKey        string
	PublicBaseURL     string
	RegistryBaseURL   string
	RegistryToken     string
	HeartbeatInterval time.Duration
	Version           string
}

// LoadConfig reads env: PORT, NEWSAPI_KEY, PUBLIC_BASE_URL, AGENTGLOBE_BASE_URL, SERVICE_REGISTRY_TOKEN, HEARTBEAT_INTERVAL.
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
	v := os.Getenv("NEWAPI_SERVER_VERSION")
	if v == "" {
		v = "0.0.0-dev"
	}
	return Config{
		Port:              p,
		NewsAPIKey:        os.Getenv("NEWSAPI_KEY"),
		PublicBaseURL:     strings.TrimRight(strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL")), "/"),
		RegistryBaseURL:   strings.TrimRight(strings.TrimSpace(os.Getenv("AGENTGLOBE_BASE_URL")), "/"),
		RegistryToken:     os.Getenv("SERVICE_REGISTRY_TOKEN"),
		HeartbeatInterval: hb,
		Version:           v,
	}
}

// NewRouter builds the Chi router. client may be nil; proxy returns 503.
func NewRouter(client *newapi.Client, publicBase, ver string, rcli *regclient.Client) *chi.Mux {
	if ver == "" {
		ver = "0.0.0-dev"
	}
	if publicBase == "" {
		publicBase = "http://127.0.0.1:8080"
	}
	r := chi.NewRouter()
	ready := client != nil
	r.Get("/health", handleHealth(ready, ver, publicBase))
	r.Get("/info", handleInfo(ready, ver, publicBase))
	r.Get("/openapi.json", handleOpenapi())
	r.Get("/capabilities", handleCapabilities())
	r.Post("/register", handleRegister(ready, rcli, publicBase, ver, svcName))
	r.Get("/v1/v1/articles", newapiProxy(ready, client, func(c *newapi.Client, ctx context.Context, q url.Values) ([]byte, error) {
		out, _, e := c.V1.Articles(ctx, q, nil)
		if e != nil {
			return nil, e
		}
		return json.Marshal(out)
	}))
	r.Get("/v1/v1/sources", newapiProxy(ready, client, func(c *newapi.Client, ctx context.Context, q url.Values) ([]byte, error) {
		out, _, e := c.V1.Sources(ctx, q, nil)
		if e != nil {
			return nil, e
		}
		return json.Marshal(out)
	}))
	r.Get("/v1/v2/top-headlines", newapiProxy(ready, client, func(c *newapi.Client, ctx context.Context, q url.Values) ([]byte, error) {
		out, _, e := c.V2.TopHeadlines(ctx, q, nil)
		if e != nil {
			return nil, e
		}
		return json.Marshal(out)
	}))
	r.Get("/v1/v2/everything", newapiProxy(ready, client, func(c *newapi.Client, ctx context.Context, q url.Values) ([]byte, error) {
		out, _, e := c.V2.Everything(ctx, q, nil)
		if e != nil {
			return nil, e
		}
		return json.Marshal(out)
	}))
	r.Get("/v1/v2/sources", newapiProxy(ready, client, func(c *newapi.Client, ctx context.Context, q url.Values) ([]byte, error) {
		out, _, e := c.V2.Sources(ctx, q, nil)
		if e != nil {
			return nil, e
		}
		return json.Marshal(out)
	}))
	return r
}

type proxyCall func(c *newapi.Client, ctx context.Context, q url.Values) ([]byte, error)

func newapiProxy(ready bool, client *newapi.Client, fn proxyCall) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !ready {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]any{"detail": "NEWSAPI_KEY not set"})
			return
		}
		if err := r.ParseForm(); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"detail": err.Error()})
			return
		}
		q := r.Form
		if q == nil {
			q = url.Values{}
		}
		b, err := fn(client, r.Context(), q)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			var ae *newapi.APIError
			if errors.As(err, &ae) {
				w.WriteHeader(http.StatusBadRequest)
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

func handleHealth(ready bool, version, publicBase string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "ok",
			"service": svcName,
			"version": version,
			"ready":   ready,
			"public_base_url": publicBase,
		})
	}
}

func handleInfo(ready bool, version, publicBase string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name":        svcName,
			"version":     version,
			"description": svcDescription,
			"ready":       ready,
			"public_base_url": publicBase,
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
		caps := []map[string]any{
			{
				"id":   "v1.articles", "method": "GET", "path": "/v1/v1/articles",
				"summary": "News API GET /v1/articles",
			},
			{
				"id":   "v1.sources", "method": "GET", "path": "/v1/v1/sources",
				"summary": "News API GET /v1/sources",
			},
			{
				"id":   "v2.top-headlines", "method": "GET", "path": "/v1/v2/top-headlines",
				"summary": "News API GET /v2/top-headlines",
			},
			{
				"id":   "v2.everything", "method": "GET", "path": "/v1/v2/everything",
				"summary": "News API GET /v2/everything",
			},
			{
				"id":   "v2.sources", "method": "GET", "path": "/v1/v2/sources",
				"summary": "News API GET /v2/sources",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"service":  svcName,
			"capabilities": caps,
		})
	}
}

func handleRegister(ready bool, rcli *regclient.Client, publicBase, version, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r
		if rcli == nil || !rcli.Capable() {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":     false,
				"reason": "AGENTGLOBE_BASE_URL and SERVICE_REGISTRY_TOKEN not both set",
			})
			return
		}
		req := regclient.RegisterRequest{
			Name:         name,
			Version:      version,
			BaseURL:      publicBase,
			Description:  svcDescription,
			Category:     "news",
			Tags:         []string{"news", "newapi", "headlines"},
			Domains:      []string{"v1.articles", "v1.sources", "v2.top-headlines", "v2.everything", "v2.sources"},
			OpenapiURL:   publicBase + "/openapi.json",
			OpenapiSpec:  nil,
			Metadata: map[string]any{
				"kind": "newapi_server", "upstream": "https://newsapi.org", "newsapi_configured": ready,
			},
		}
		if err := rcli.Register(r.Context(), req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(map[string]string{"detail": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}
}

// RunContext starts the HTTP server and a heartbeat goroutine. Blocks until ctx is done or the server returns.
func RunContext(ctx context.Context, cfg Config, rcli *regclient.Client, out io.Writer) error {
	if out == nil {
		out = os.Stderr
	}
	var c *newapi.Client
	if strings.TrimSpace(cfg.NewsAPIKey) != "" {
		var err error
		c, err = newapi.New(cfg.NewsAPIKey, nil)
		if err != nil {
			return err
		}
	} else {
		_, _ = io.WriteString(out, "httpserver: NEWSAPI_KEY is empty: proxy returns 503; health still OK\n")
	}
	pub := cfg.PublicBaseURL
	if pub == "" {
		pub = fmt.Sprintf("http://127.0.0.1:%d", cfg.Port)
	}
	h := NewRouter(c, pub, cfg.Version, rcli)
	if rcli != nil && rcli.Capable() {
		_, _ = io.WriteString(out, "httpserver: sending initial registry POST\n")
		go func() {
			req := regclient.RegisterRequest{
				Name:         svcName,
				Version:      cfg.Version,
				BaseURL:      pub,
				Description:  svcDescription,
				Category:     "news",
				Tags:         []string{"news", "newapi", "headlines"},
				Domains:      []string{"v1.articles", "v1.sources", "v2.top-headlines", "v2.everything", "v2.sources"},
				OpenapiURL:   pub + "/openapi.json",
				Metadata: map[string]any{
					"kind": "newapi_server", "upstream": "https://newsapi.org", "newsapi_configured": c != nil,
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
					_ = rcli.Heartbeat(context.Background(), svcName, pub, "active")
				}
			}
		}()
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
