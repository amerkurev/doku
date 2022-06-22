package handler

import (
	"net/http"
)

// ServeFile handles requests for a specific file..
func ServeFile(filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filename)
	}
}
