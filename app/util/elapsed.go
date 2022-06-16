package util

import (
	"time"

	log "github.com/sirupsen/logrus"
)

// Elapsed measures execution time.
func Elapsed(what string) func() {
	start := time.Now()
	return func() {
		log.WithField("took", time.Since(start)).Debug(what)
	}
}
