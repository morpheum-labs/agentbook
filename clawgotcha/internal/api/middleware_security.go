package api

import (
	"context"
	"crypto/subtle"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/morpheumlabs/agentbook/clawgotcha/internal/httperr"
	"golang.org/x/time/rate"
)

func clientIP(r *http.Request) string {
	if x := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); x != "" {
		if i := strings.IndexByte(x, ','); i >= 0 {
			x = strings.TrimSpace(x[:i])
		}
		return x
	}
	if x := strings.TrimSpace(r.Header.Get("X-Real-Ip")); x != "" {
		return x
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (s *Server) requireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimSpace(s.apiKey)
		if key == "" {
			next.ServeHTTP(w, r)
			return
		}
		tok := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if tok == "" {
			tok = strings.TrimSpace(r.Header.Get("X-API-Key"))
		}
		if tok == "" || subtle.ConstantTimeCompare([]byte(tok), []byte(key)) != 1 {
			httperr.Write(w, r, httperr.Forbidden("invalid or missing API key"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func maxBodyBytes(n int64) func(http.Handler) http.Handler {
	if n <= 0 {
		n = 1 << 20
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > n {
				httperr.Write(w, r, httperr.PayloadTooLarge("request body too large"))
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, n)
			next.ServeHTTP(w, r)
		})
	}
}

type ipRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*visitorLimiter
	limit    rate.Limit
	burst    int
	ttl      time.Duration
}

type visitorLimiter struct {
	lim      *rate.Limiter
	lastSeen time.Time
}

func newIPRateLimiter(rps float64, burst int, ttl time.Duration) *ipRateLimiter {
	if burst < 1 {
		burst = 1
	}
	return &ipRateLimiter{
		limiters: make(map[string]*visitorLimiter),
		limit:    rate.Limit(rps),
		burst:    burst,
		ttl:      ttl,
	}
}

func (irl *ipRateLimiter) Allow(ip string) bool {
	now := time.Now().UTC()
	irl.mu.Lock()
	defer irl.mu.Unlock()
	v, ok := irl.limiters[ip]
	if !ok {
		v = &visitorLimiter{lim: rate.NewLimiter(irl.limit, irl.burst)}
		irl.limiters[ip] = v
	}
	v.lastSeen = now
	return v.lim.Allow()
}

func (irl *ipRateLimiter) cleanupLoop(ctx context.Context) {
	t := time.NewTicker(2 * time.Minute)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			cutoff := time.Now().UTC().Add(-irl.ttl)
			irl.mu.Lock()
			for ip, v := range irl.limiters {
				if v.lastSeen.Before(cutoff) {
					delete(irl.limiters, ip)
				}
			}
			irl.mu.Unlock()
		}
	}
}

func (irl *ipRateLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !irl.Allow(ip) {
			httperr.Write(w, r, httperr.TooManyRequests("rate limit exceeded"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
