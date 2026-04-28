package api

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "clawgotcha",
			Name:      "http_requests_total",
			Help:      "HTTP requests processed",
		},
		[]string{"method", "code"},
	)
	httpDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "clawgotcha",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request latency",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method"},
	)
)

func slogRequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		reqID := middleware.GetReqID(r.Context())
		code := ww.Status()
		if code == 0 {
			code = http.StatusOK
		}
		d := time.Since(start).Seconds()
		httpRequests.WithLabelValues(r.Method, strconv.Itoa(code)).Inc()
		httpDuration.WithLabelValues(r.Method).Observe(d)
		slog.Info("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", code,
			"bytes", ww.BytesWritten(),
			"dur_ms", time.Since(start).Milliseconds(),
			"req_id", reqID,
		)
	})
}
