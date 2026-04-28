package api

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/morpheumlabs/agentbook/clawgotcha/internal/httperr"
)

func (s *Server) requireInternalToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokCfg := strings.TrimSpace(s.internalToken)
		if tokCfg == "" {
			httperr.Write(w, r, httperr.ServiceUnavailable("CLAWGOTCHA_INTERNAL_TOKEN is not set"))
			return
		}
		tok := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if tok == "" {
			tok = strings.TrimSpace(r.Header.Get("X-Internal-Token"))
		}
		if tok == "" || subtle.ConstantTimeCompare([]byte(tok), []byte(tokCfg)) != 1 {
			httperr.Write(w, r, httperr.Forbidden("invalid or missing token"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
