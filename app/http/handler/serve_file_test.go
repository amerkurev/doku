package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ServeFile(t *testing.T) {
	fh, err := os.CreateTemp(t.TempDir(), "doku_serve_file_*")
	require.NoError(t, err)
	defer fh.Close()

	s := "some content"
	n, err := fh.WriteString(s)
	require.NoError(t, err)
	require.Equal(t, len(s), n)

	wr := httptest.NewRecorder()
	handler := ServeFile(fh.Name())

	req, err := http.NewRequest("GET", "http://example.com", http.NoBody)
	require.NoError(t, err)

	handler.ServeHTTP(wr, req)
	require.Equal(t, http.StatusOK, wr.Code)

	b, err := io.ReadAll(wr.Body)
	assert.NoError(t, err)
	assert.Equal(t, s, string(b))
}
