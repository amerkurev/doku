package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/amerkurev/doku/app/store"
)

func Test_ContentTypeJSON(t *testing.T) {
	wr := httptest.NewRecorder()
	handler := ContentTypeJSON(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req, err := http.NewRequest("GET", "http://example.com", http.NoBody)
	require.NoError(t, err)

	handler.ServeHTTP(wr, req)
	assert.Equal(t, "application/json; charset=utf-8", wr.Result().Header.Get("Content-Type"))
}

func Test_LongPolling(t *testing.T) {
	err := store.Initialize()
	require.NoError(t, err)

	wr := httptest.NewRecorder()
	handler := LongPolling(time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	start := time.Now()
	go func() {
		time.Sleep(time.Millisecond * 100)
		store.NotifyAll()
	}()
	req, err := http.NewRequest("GET", "http://example.com", http.NoBody)
	require.NoError(t, err)

	handler.ServeHTTP(wr, req)
	assert.Less(t, time.Since(start), time.Second)
}

func Test_BasicAuthentication(t *testing.T) {
	allowed := []string{
		"test:$2y$05$zMxDmK65SjcH2vJQNopVSO/nE8ngVLx65RoETyHpez7yTS/8CLEiW",
		"test2:$2y$05$TLQqHh6VT4JxysdKGPOlJeSkkMsv.Ku/G45i7ssIm80XuouCrES12 ",
		"bad bad",
	}

	client := http.Client{}

	tbl := []struct {
		reqFn func(r *http.Request)
		ok    bool
	}{
		{func(r *http.Request) {}, false},
		{func(r *http.Request) { r.SetBasicAuth("test", "passwd") }, true},
		{func(r *http.Request) { r.SetBasicAuth("test", "passwdbad") }, false},
		{func(r *http.Request) { r.SetBasicAuth("test2", "passwd2") }, true},
		{func(r *http.Request) { r.SetBasicAuth("test2", "passwbad") }, false},
		{func(r *http.Request) { r.SetBasicAuth("testbad", "passwbad") }, false},
	}

	handler := BasicAuthentication(allowed)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ts := httptest.NewServer(handler)
	for i, tt := range tbl {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			req, err := http.NewRequest("GET", ts.URL, http.NoBody)
			require.NoError(t, err)
			tt.reqFn(req)
			resp, err := client.Do(req)
			require.NoError(t, err)
			err = resp.Body.Close()
			require.NoError(t, err)
			if tt.ok {
				require.Equal(t, http.StatusOK, resp.StatusCode)
				return
			}
			require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	}
}

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
		require.Equal(t, len(lines), 3)

		logReq := lines[0]
		assert.True(t, strings.Contains(logReq, "level=debug"))
		assert.True(t, strings.Contains(logReq, "request started"))
		assert.True(t, strings.Contains(logReq, "http_method=GET"))
		assert.True(t, strings.Contains(logReq, "http_proto=HTTP/1.1"))
		assert.True(t, strings.Contains(logReq, "http_scheme=http"))
		assert.True(t, strings.Contains(logReq, "remote_addr="))
		assert.True(t, strings.Contains(logReq, "uri=\"http://example.com\""))
		assert.True(t, strings.Contains(logReq, "user_agent=doku-test"))

		logRes := lines[1]
		assert.True(t, strings.Contains(logRes, "level=debug"))
		assert.True(t, strings.Contains(logRes, "request complete"))
		assert.True(t, strings.Contains(logRes, "http_method=GET"))
		assert.True(t, strings.Contains(logRes, "http_proto=HTTP/1.1"))
		assert.True(t, strings.Contains(logRes, "http_scheme=http"))
		assert.True(t, strings.Contains(logRes, "remote_addr="))
		assert.True(t, strings.Contains(logRes, "resp_bytes_length=0"))
		assert.True(t, strings.Contains(logRes, "resp_elapsed_ms="))
		assert.True(t, strings.Contains(logRes, "resp_status=0"))
		assert.True(t, strings.Contains(logRes, "uri=\"http://example.com\""))
		assert.True(t, strings.Contains(logRes, "user_agent=doku-test"))
	})

	t.Run("panic", func(t *testing.T) {
		l := StructuredLogger{log.StandardLogger()}
		entry := l.NewLogEntry(req)
		entry.Panic("panic!", []byte{})
	})
}
