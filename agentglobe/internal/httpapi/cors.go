package httpapi

import (
	"net/http"
	"strings"
)

func originAllowed(allowed []string, origin string) bool {
	origin = strings.TrimSpace(strings.TrimRight(origin, "/"))
	if origin == "" {
		return false
	}
	for _, a := range allowed {
		if a == origin {
			return true
		}
	}
	return false
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowed := s.Cfg.CORSAllowedOrigins
		origin := r.Header.Get("Origin")

		if len(allowed) == 0 {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin != "" && originAllowed(allowed, origin) {
			w.Header().Set("Access-Control-Allow-Origin", strings.TrimSpace(strings.TrimRight(origin, "/")))
			w.Header().Set("Vary", "Origin")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// wsCheckOrigin mirrors CORS policy for browser WebSocket handshakes. Non-browser clients often omit Origin.
func (s *Server) wsCheckOrigin(r *http.Request) bool {
	allowed := s.Cfg.CORSAllowedOrigins
	if len(allowed) == 0 {
		return true
	}
	o := r.Header.Get("Origin")
	if o == "" {
		return true
	}
	return originAllowed(allowed, o)
}
