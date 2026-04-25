package worldmon

import (
	"encoding/json"
	"strings"
)

// ErrorBody is the common JSON shape for 4xx/5xx responses from the edge gateway
// ([server/gateway.ts] and [server/error-mapper.ts]).
//
// [server/gateway.ts]: https://github.com/koala73/worldmonitor/blob/main/server/gateway.ts
// [server/error-mapper.ts]: https://github.com/koala73/worldmonitor/blob/main/server/error-mapper.ts
type ErrorBody struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	// Debug is optional; gateway may include _debug for key validation in some modes.
	Debug interface{} `json:"_debug,omitempty"`
	// RetryAfter may appear on 429 (seconds).
	RetryAfter *int `json:"retryAfter,omitempty"`
}

// ParseErrorBody decodes a JSON error body, preferring the "error" field then "message".
// It returns the empty string if neither is set.
func ParseErrorBody(body []byte) string {
	var b ErrorBody
	if err := json.Unmarshal(body, &b); err != nil {
		return strings.TrimSpace(string(body))
	}
	if s := strings.TrimSpace(b.Error); s != "" {
		return s
	}
	if s := strings.TrimSpace(b.Message); s != "" {
		return s
	}
	return ""
}
