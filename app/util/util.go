package util

import (
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

// PrintExecTime measures execution time.
func PrintExecTime(what string) func() {
	start := time.Now()
	return func() {
		log.WithField("took", time.Since(start)).Debug(what)
	}
}

// DirSize returns directory total size.
func DirSize(path string) (int64, int64, error) {
	var size int64
	var files int64
	err := filepath.Walk(path, func(_ string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !fi.IsDir() {
			size += fi.Size()
			files++
		}
		return err
	})
	return size, files, err
}
