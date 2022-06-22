package handler

import (
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
)

// SinglePageApplication renders and serves the index.html.
func SinglePageApplication(indexHTML, title, header string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles(indexHTML)
		if err != nil {
			log.WithField("err", err).Error("failed to parse template")
			http.ServeFile(w, r, indexHTML)
			return
		}
		w.Header().Set("Content-Type", "text/html")

		err = t.Execute(w, &struct {
			Title  string
			Header string
		}{
			Title:  title,
			Header: header,
		})

		if err != nil {
			log.WithField("err", err).Error("failed to apply a parsed template to the specified data")
			http.ServeFile(w, r, indexHTML)
			return
		}
	}
}
