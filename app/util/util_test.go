package util

import (
	"bytes"
	"errors"
	"os"
	"path"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_DirSize(t *testing.T) {
	p := t.TempDir()

	d := []byte("Hello, world!")
	err := os.WriteFile(path.Join(p, "greeting.txt"), d, 0644)
	require.NoError(t, err)

	size, files, err := DirSize(p)
	require.NoError(t, err)
	assert.Equal(t, files, int64(1))
	assert.Equal(t, size, int64(len(d)))

	size, files, err = DirSize("/the-wrong-path")
	assert.True(t, errors.Is(err, os.ErrNotExist))
	assert.Equal(t, files, int64(0))
	assert.Equal(t, size, int64(0))
}

func Test_PrintExecTime(t *testing.T) {
	var buf bytes.Buffer
	log.StandardLogger().Out = &buf
	log.SetLevel(log.DebugLevel)
	PrintExecTime("text")()

	assert.True(t, strings.Contains(buf.String(), "level=debug"))
	assert.True(t, strings.Contains(buf.String(), "msg=text"))
	assert.True(t, strings.Contains(buf.String(), "took="))
	assert.True(t, strings.Contains(buf.String(), "time="))
}
