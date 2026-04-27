package httperr

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"gorm.io/gorm"
)

// PublicError is returned to the client (safe to expose).
type PublicError struct {
	Code   string `json:"code"`
	Detail string `json:"detail,omitempty"`
}

func (e *PublicError) Error() string { return e.Code + ": " + e.Detail }

// BadRequest wraps a client error.
func BadRequest(code string, err error) *PublicError {
	pe := &PublicError{Code: code, Detail: code}
	if err != nil {
		pe.Detail = err.Error()
	}
	return pe
}

// NotFound is a 404 public error.
func NotFound(what string) *PublicError {
	return &PublicError{Code: "not_found", Detail: what + " not found"}
}

// Write serializes err as JSON. Logs internal errors.
func Write(w http.ResponseWriter, r *http.Request, err error) {
	var pe *PublicError
	switch {
	case errors.As(err, &pe):
		status := http.StatusBadRequest
		if pe.Code == "not_found" {
			status = http.StatusNotFound
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": pe})
	case errors.Is(err, gorm.ErrRecordNotFound):
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": &PublicError{Code: "not_found", Detail: "record not found"},
		})
	default:
		slog.Error("internal error", "err", err, "path", r.URL.Path)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": &PublicError{Code: "internal_error", Detail: "internal error"},
		})
	}
}
