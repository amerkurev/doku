package handler

import (
	"fmt"
	"net/http"
)

// FavIcon handles requests for /favicon.ico.
func FavIcon(staticFolder string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(staticFolder + "/favicon.ico")
		http.ServeFile(w, r, staticFolder+"/favicon.ico")
	}
}
