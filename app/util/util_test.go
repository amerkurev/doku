package util

import (
	"bytes"
	"io/fs"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_DirSize(t *testing.T) {
	fh, err := os.CreateTemp(t.TempDir(), "doku_dir_size_*")
	require.NoError(t, err)
	defer fh.Close()

	s := "some content"
	n, err := fh.WriteString(s)
	require.NoError(t, err)
	require.Equal(t, len(s), n)

	size, files, err := DirSize(fh.Name())
	require.NoError(t, err)
	assert.Equal(t, int64(1), files)
	assert.Equal(t, int64(len(s)), size)

	size, files, err = DirSize("/the-wrong-path")
	var perr *fs.PathError
	assert.ErrorAs(t, err, &perr)
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
