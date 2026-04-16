package httpapi

import (
	"os"
	"strings"
	"time"
)

// handlerRequestTimeout returns per-request context deadline for /api/v1 routes (excluding WebSocket).
// Set HTTP_HANDLER_TIMEOUT (Go duration, e.g. "90s", "3m"); "0" or "off" disables chi Timeout middleware.
func handlerRequestTimeout() time.Duration {
	s := strings.TrimSpace(os.Getenv("HTTP_HANDLER_TIMEOUT"))
	if s == "" {
		return 2 * time.Minute
	}
	if s == "0" || strings.EqualFold(s, "off") {
		return 0
	}
	d, err := time.ParseDuration(s)
	if err != nil || d <= 0 {
		return 2 * time.Minute
	}
	return d
}
