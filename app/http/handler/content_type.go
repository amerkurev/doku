package handler

import "net/http"

// ContentTypeJSON sets Content-Type (HTTP header) to `application/json`.
func ContentTypeJSON(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
