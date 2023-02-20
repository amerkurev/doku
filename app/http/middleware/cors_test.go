package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_DevCORS(t *testing.T) {
	wr := httptest.NewRecorder()
	handler := DevCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req, err := http.NewRequest("GET", "http://example.com", http.NoBody)
	require.NoError(t, err)

	handler.ServeHTTP(wr, req)
	assert.Equal(t, "true", wr.Result().Header.Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "http://localhost:3000", wr.Result().Header.Get("Access-Control-Allow-Origin"))

	req, err = http.NewRequest("OPTIONS", "http://example.com", http.NoBody)
	require.NoError(t, err)
	handler.ServeHTTP(wr, req)
	assert.Equal(t, "true", wr.Result().Header.Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "http://localhost:3000", wr.Result().Header.Get("Access-Control-Allow-Origin"))
}
