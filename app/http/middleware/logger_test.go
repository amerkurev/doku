package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewStructuredLogger(t *testing.T) {
	wr := httptest.NewRecorder()
	handler := NewStructuredLogger(log.StandardLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req, err := http.NewRequest("GET", "http://example.com", bytes.NewBufferString("123456"))
	require.NoError(t, err)
	req.Header.Set("User-Agent", "doku-test")

	t.Run("info", func(t *testing.T) {
		var buf bytes.Buffer
		log.StandardLogger().Out = &buf
		log.SetLevel(log.InfoLevel)

		handler.ServeHTTP(wr, req)
		assert.Equal(t, "", buf.String())
	})

	t.Run("debug", func(t *testing.T) {
		var buf bytes.Buffer
		log.StandardLogger().Out = &buf
		log.SetLevel(log.DebugLevel)

		handler.ServeHTTP(wr, req)
		lines := strings.Split(buf.String(), "\n")
		require.Equal(t, 3, len(lines))

		logReq := lines[0]
		assert.Contains(t, logReq, "level=debug")
		assert.Contains(t, logReq, "request started")
		assert.Contains(t, logReq, "http_method=GET")
		assert.Contains(t, logReq, "http_proto=HTTP/1.1")
		assert.Contains(t, logReq, "http_scheme=http")
		assert.Contains(t, logReq, "remote_addr=")
		assert.Contains(t, logReq, "uri=\"http://example.com\"")
		assert.Contains(t, logReq, "user_agent=doku-test")

		logRes := lines[1]
		assert.Contains(t, logRes, "level=debug")
		assert.Contains(t, logRes, "request complete")
		assert.Contains(t, logRes, "http_method=GET")
		assert.Contains(t, logRes, "http_proto=HTTP/1.1")
		assert.Contains(t, logRes, "http_scheme=http")
		assert.Contains(t, logRes, "remote_addr=")
		assert.Contains(t, logRes, "resp_bytes_length=0")
		assert.Contains(t, logRes, "resp_elapsed_ms=")
		assert.Contains(t, logRes, "resp_status=0")
		assert.Contains(t, logRes, "uri=\"http://example.com\"")
		assert.Contains(t, logRes, "user_agent=doku-test")
	})

	t.Run("panic", func(t *testing.T) {
		l := StructuredLogger{log.StandardLogger()}
		entry := l.NewLogEntry(req)
		entry.Panic("panic!", []byte{})
	})
}
