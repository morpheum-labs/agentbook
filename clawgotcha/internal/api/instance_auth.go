package api

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/db"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/httperr"
)

type runtimeInstanceCtxKey struct{}

// WithRuntimeInstance attaches the authenticated runtime instance to ctx.
func WithRuntimeInstance(ctx context.Context, inst *db.SwarmRuntimeInstance) context.Context {
	return context.WithValue(ctx, runtimeInstanceCtxKey{}, inst)
}

// RuntimeInstanceFromContext returns the instance set by requireInstanceAuth, or nil.
func RuntimeInstanceFromContext(ctx context.Context) *db.SwarmRuntimeInstance {
	v, _ := ctx.Value(runtimeInstanceCtxKey{}).(*db.SwarmRuntimeInstance)
	return v
}

func hashInstanceAPISecretHex(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

// requireInstanceAuth verifies X-Instance-Secret or Authorization: Bearer against the row for {instance_name}.
func (s *Server) requireInstanceAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimSpace(chi.URLParam(r, "instance_name"))
		if name == "" {
			httperr.Write(w, r, httperr.BadRequest("instance_name required", nil))
			return
		}
		var inst db.SwarmRuntimeInstance
		if err := s.db.Where("instance_name = ?", name).First(&inst).Error; err != nil {
			httperr.Write(w, r, httperr.Forbidden("invalid instance or secret"))
			return
		}
		tok := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if tok == "" {
			tok = strings.TrimSpace(r.Header.Get("X-Instance-Secret"))
		}
		if tok == "" || strings.TrimSpace(inst.ApiSecretHash) == "" {
			httperr.Write(w, r, httperr.Forbidden("invalid instance or secret"))
			return
		}
		want := strings.TrimSpace(inst.ApiSecretHash)
		got := hashInstanceAPISecretHex(tok)
		if len(got) != len(want) || subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
			httperr.Write(w, r, httperr.Forbidden("invalid instance or secret"))
			return
		}
		next.ServeHTTP(w, r.WithContext(WithRuntimeInstance(r.Context(), &inst)))
	})
}
