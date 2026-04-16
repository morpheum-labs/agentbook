package httpapi

import (
	"context"
	"errors"
	"log"
	"net/http"

	"gorm.io/gorm"
)

type gormDBCtxKey struct{}

// RequestDB returns the per-request Gorm handle installed by requestDBMiddleware, or nil if missing.
func RequestDB(r *http.Request) *gorm.DB {
	if r == nil {
		return nil
	}
	db, _ := r.Context().Value(gormDBCtxKey{}).(*gorm.DB)
	return db
}

func (s *Server) requestDBMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gdb := s.DB.WithContext(r.Context())
		ctx := context.WithValue(r.Context(), gormDBCtxKey{}, gdb)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
		// Low-noise operator signal when the request deadline fired (e.g. chi Timeout or client abort).
		if err := r.Context().Err(); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				log.Printf("httpapi: request deadline exceeded %s %s", r.Method, r.URL.Path)
			} else if errors.Is(err, context.Canceled) {
				log.Printf("httpapi: request canceled %s %s", r.Method, r.URL.Path)
			}
		}
	})
}
