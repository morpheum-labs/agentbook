package worldmon

import (
	"encoding/json"
	"strings"
)

// ErrorBody is a common JSON shape for 4xx/5xx responses: optional "error", "message",
// "_debug", and "retryAfter" fields.
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
