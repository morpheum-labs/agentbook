package mcp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

// Unit tests: mock 6551-style routes and assert query forwarding (get_hot_news uses
// category; optional subcategory and limit). RSS/worldmon list-feed-digest does not
// have API “categories” — you pass feed URLs in ?feeds=.

func TestGetHotNews_ForwardsCategoryAndSubLimit(t *testing.T) {
	t.Parallel()
	var last string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open/free_hot" {
			http.NotFound(w, r)
			return
		}
		last = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"ok":true,"from":"test"}`)
	}))
	t.Cleanup(srv.Close)

	s := &State{
		DailyNewsAPIBase: srv.URL,
		HTTPClient:       srv.Client(),
	}
	_, err := s.getHotNews(context.Background(), GetHotNewsArgs{
		Category:    "ai",
		Subcategory: "chips",
		Limit:       7,
	})
	if err != nil {
		t.Fatal(err)
	}
	q, err := url.ParseQuery(last)
	if err != nil {
		t.Fatal(err)
	}
	if q.Get("category") != "ai" {
		t.Fatalf("category: %q", last)
	}
	if q.Get("subcategory") != "chips" {
		t.Fatalf("subcategory: %q", last)
	}
	if q.Get("limit") != "7" {
		t.Fatalf("limit: %q", last)
	}
}

func TestGetHotNews_CategoryRequired(t *testing.T) {
	t.Parallel()
	s := &State{DailyNewsAPIBase: "http://nope", HTTPClient: &http.Client{}}
	_, err := s.getHotNews(context.Background(), GetHotNewsArgs{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetNewsCategories_UsesPath(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open/free_categories" {
			http.NotFound(w, r)
			return
		}
		_, _ = io.WriteString(w, `{"success":true,"categories":[]}`)
	}))
	t.Cleanup(srv.Close)
	s := &State{DailyNewsAPIBase: srv.URL, HTTPClient: srv.Client()}
	_, err := s.getNewsCategories(context.Background(), GetNewsCategoriesArgs{})
	if err != nil {
		t.Fatal(err)
	}
}

// TestIntegrationDailyNewsEveryCategory calls the public daily-news API: list
// categories, then for each top-level and each (parent, sub) call free_hot.
// Run:  cd agentglobe && DAILY_NEWS_INTEGRATION=1 GOWORK=off go test -v -count=1 -run TestIntegrationDailyNewsEveryCategory ./internal/mcp
func TestIntegrationDailyNewsEveryCategory(t *testing.T) {
	if os.Getenv("DAILY_NEWS_INTEGRATION") == "" {
		t.Skip("set DAILY_NEWS_INTEGRATION=1 to hit the real 6551 daily-news API (network)")
	}
	base := strings.TrimRight(os.Getenv("DAILY_NEWS_API_BASE"), "/")
	if base == "" {
		base = DefaultDailyNewsAPIBase
	}
	client := &http.Client{Timeout: 60 * time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	t.Cleanup(cancel)

	catURL := base + "/open/free_categories"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, catURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("free_categories: %v", err)
	}
	b, rerr := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if rerr != nil {
		t.Fatal(rerr)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		t.Fatalf("free_categories http %d: %s", res.StatusCode, string(b))
	}
	if !json.Valid(b) {
		t.Fatal("free_categories: invalid json")
	}
	var payload struct {
		Success    bool `json:"success"`
		Categories []struct {
			Key           string `json:"key"`
			Subcategories []struct {
				Key string `json:"key"`
			} `json:"subcategories"`
		} `json:"categories"`
	}
	if err := json.Unmarshal(b, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Categories) == 0 {
		t.Fatal("no categories in response; schema change?")
	}
	hot := func(t *testing.T, cat, sub string) {
		t.Helper()
		u, _ := url.Parse(base + "/open/free_hot")
		q := u.Query()
		q.Set("category", cat)
		if sub != "" {
			q.Set("subcategory", sub)
		}
		q.Set("limit", "2")
		u.RawQuery = q.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			t.Fatalf("%s: %v", u.String(), err)
		}
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("%s: %v", u.String(), err)
		}
		body, _ := io.ReadAll(res.Body)
		_ = res.Body.Close()
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			t.Errorf("free_hot %s&… http %d: %s", u.Path, res.StatusCode, string(body))
			return
		}
		if !json.Valid(body) {
			t.Errorf("free_hot category=%q sub=%q: invalid json", cat, sub)
		}
	}
	for _, c := range payload.Categories {
		cat := c.Key
		if cat == "" {
			continue
		}
		t.Run("top_"+url.PathEscape(cat), func(t *testing.T) {
			t.Parallel()
			hot(t, cat, "")
		})
	}
	for _, c := range payload.Categories {
		parent := c.Key
		if parent == "" {
			continue
		}
		for _, s := range c.Subcategories {
			k := s.Key
			if k == "" {
				continue
			}
			pair := parent + "__" + k
			t.Run("sub_"+url.PathEscape(pair), func(t *testing.T) {
				t.Parallel()
				hot(t, parent, k)
			})
		}
	}
}
