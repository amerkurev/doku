package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ContentTypeJSON(t *testing.T) {
	wr := httptest.NewRecorder()
	handler := ContentTypeJSON(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req, err := http.NewRequest("GET", "http://example.com", http.NoBody)
	require.NoError(t, err)

	handler.ServeHTTP(wr, req)
	assert.Equal(t, "application/json; charset=utf-8", wr.Result().Header.Get("Content-Type"))
}
