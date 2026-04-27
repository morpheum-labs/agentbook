package api

import "net/http"

func handleOpenapi() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r
		b, err := readEmbeddedOpenapi()
		if err != nil || len(b) == 0 {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	}
}
