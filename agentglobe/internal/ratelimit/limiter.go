package ratelimit

import (
	"fmt"
	"sync"
	"time"

	"github.com/morpheumlabs/agentbook/agentglobe/internal/config"
)

// Limiter is a sliding-window per-key limiter (matches minibook ratelimit.py).
type Limiter struct {
	mu      sync.Mutex
	history map[string][]entry // key -> timestamps + action
	limits  map[string]limit
}

type entry struct {
	ts     float64
	action string
}

type limit struct {
	max    int
	window int
}

var defaultLimits = map[string]limit{
	"post":       {10, 60},
	"comment":    {60, 60},
	"register":   {5, 3600},
	"attachment": {30, 3600},
}

func New(cfg *config.Config) *Limiter {
	l := &Limiter{
		history: make(map[string][]entry),
		limits:  make(map[string]limit),
	}
	for k, v := range defaultLimits {
		l.limits[k] = v
	}
	if cfg != nil && cfg.RateLimits != nil {
		for action, s := range cfg.RateLimits {
			def := defaultLimits[action]
			maxC := s.Limit
			if maxC == 0 {
				maxC = def.max
			}
			win := s.Window
			if win == 0 {
				win = def.window
			}
			l.limits[action] = limit{max: maxC, window: win}
		}
	}
	return l
}

func (l *Limiter) cleanup(key string, window int) {
	cutoff := time.Now().UnixNano()/1e9 - int64(window)
	kept := l.history[key][:0]
	for _, e := range l.history[key] {
		if int64(e.ts) > cutoff {
			kept = append(kept, e)
		}
	}
	l.history[key] = kept
}

func (l *Limiter) retryAfter(key, action string, window int) int {
	now := float64(time.Now().UnixNano()) / 1e9
	cutoff := now - float64(window)
	var oldest *float64
	for _, e := range l.history[key] {
		if e.action != action || e.ts <= cutoff {
			continue
		}
		if oldest == nil || e.ts < *oldest {
			t := e.ts
			oldest = &t
		}
	}
	if oldest == nil {
		return 1
	}
	sec := int((*oldest + float64(window)) - now)
	if sec < 1 {
		return 1
	}
	return sec
}

// Check returns (retryAfter, error) — error is rate limit message; retryAfter for header.
func (l *Limiter) Check(key, action string) (int, error) {
	lim, ok := l.limits[action]
	if !ok {
		return 0, nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cleanup(key, lim.window)
	count := 0
	for _, e := range l.history[key] {
		if e.action == action {
			count++
		}
	}
	if count >= lim.max {
		ra := l.retryAfter(key, action, lim.window)
		return ra, fmt.Errorf("Rate limit exceeded: max %d %ss per %ds", lim.max, action, lim.window)
	}
	l.history[key] = append(l.history[key], entry{ts: float64(time.Now().UnixNano()) / 1e9, action: action})
	return 0, nil
}

func (l *Limiter) Stats(key string) map[string]any {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := float64(time.Now().UnixNano()) / 1e9
	out := make(map[string]any)
	for action, lim := range l.limits {
		cutoff := now - float64(lim.window)
		used := 0
		var oldestIn *float64
		for _, e := range l.history[key] {
			if e.action != action || e.ts <= cutoff {
				continue
			}
			used++
			if oldestIn == nil || e.ts < *oldestIn {
				t := e.ts
				oldestIn = &t
			}
		}
		resetIn := lim.window
		if oldestIn != nil {
			resetIn = int((*oldestIn + float64(lim.window)) - now)
			if resetIn < 0 {
				resetIn = 0
			}
		}
		rem := lim.max - used
		if rem < 0 {
			rem = 0
		}
		out[action] = map[string]any{
			"used":             used,
			"limit":            lim.max,
			"window_seconds":   lim.window,
			"remaining":        rem,
			"reset_in_seconds": resetIn,
		}
	}
	return out
}
