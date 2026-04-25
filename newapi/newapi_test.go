package newapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestNew_RequiresKey(t *testing.T) {
	_, err := New("   ")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestClient_requestURL_cors(t *testing.T) {
	c, err := New("k", WithBaseURL("https://newsapi.org"), WithCORSProxyURL("https://proxy.example/"))
	if err != nil {
		t.Fatal(err)
	}
	got := c.requestURL("/v2/top-headlines", url.Values{"a": []string{"1"}})
	want := "https://proxy.example/https://newsapi.org/v2/top-headlines?a=1"
	if got != want {
		t.Fatalf("url:\n want %q\n  got %q", want, got)
	}
}

func TestV2_TopHeadlines_NilParams_DefaultLanguage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/top-headlines" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		if q.Get("language") != "en" {
			t.Errorf("default language: got %q", q.Get("language"))
		}
		_, _ = w.Write([]byte(`{"status":"ok","totalResults":0,"articles":[]}`))
	}))
	t.Cleanup(srv.Close)
	c, err := New("test-key", WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = c.V2.TopHeadlines(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestV2_ErrorBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"status":"error","code":"paramInvalid","message":"nope"}`))
	}))
	t.Cleanup(srv.Close)
	c, err := New("k", WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = c.V2.Everything(context.Background(), url.Values{"q": []string{"x"}})
	if err == nil {
		t.Fatal("expected error")
	}
	var ae *APIError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
}

func TestV1_Sources_NoAPIKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "" {
			t.Error("v1 /sources should not send X-Api-Key (matches Node client)")
		}
		_, _ = w.Write([]byte(`{"status":"ok","sources":[]}`))
	}))
	t.Cleanup(srv.Close)
	c, err := New("k", WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = c.V1.Sources(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestV1_Articles_SendsKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "k" {
			t.Error("v1 /articles should send key")
		}
		_, _ = w.Write([]byte(`{"status":"ok","source":"x","sortBy":"y","articles":[]}`))
	}))
	t.Cleanup(srv.Close)
	c, err := New("k", WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	u := url.Values{"source": []string{"x"}}
	_, _, err = c.V1.Articles(context.Background(), u)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWithNoCache_SetsHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-No-Cache") != "true" {
			t.Error("expected X-No-Cache: true")
		}
		_, _ = w.Write([]byte(`{"status":"ok","articles":[]}`))
	}))
	t.Cleanup(srv.Close)
	c, err := New("k", WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	u := url.Values{"q": []string{"a"}}
	_, _, err = c.V2.Everything(context.Background(), u, WithNoCache())
	if err != nil {
		t.Fatal(err)
	}
}

func TestHTTPNon2XX(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	t.Cleanup(srv.Close)
	c, err := New("k", WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = c.V2.Sources(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "500") {
		t.Fatalf("expected HTTP 500 error, got %v", err)
	}
}
