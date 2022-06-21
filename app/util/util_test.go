package util

import (
	"bytes"
	"errors"
	"os"
	"path"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_DirSize(t *testing.T) {
	p := t.TempDir()

	d := []byte("some content")
	err := os.WriteFile(path.Join(p, "greeting.txt"), d, 0644)
	require.NoError(t, err)

	size, files, err := DirSize(p)
	require.NoError(t, err)
	assert.Equal(t, int64(1), files)
	assert.Equal(t, int64(len(d)), size)

	size, files, err = DirSize("/the-wrong-path")
	assert.True(t, errors.Is(err, os.ErrNotExist))
	assert.Equal(t, int64(0), files)
	assert.Equal(t, int64(0), size)
}

func Test_PrintExecTime(t *testing.T) {
	var buf bytes.Buffer
	log.StandardLogger().Out = &buf
	log.SetLevel(log.DebugLevel)
	PrintExecTime("text")()

	assert.Contains(t, buf.String(), "level=debug")
	assert.Contains(t, buf.String(), "msg=text")
	assert.Contains(t, buf.String(), "took=")
	assert.Contains(t, buf.String(), "time=")
}
