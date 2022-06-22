package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_FavIcon(t *testing.T) {
	wr := httptest.NewRecorder()
	handler := FavIcon("../../../frontend/static")

	req, err := http.NewRequest("GET", "http://example.com", http.NoBody)
	require.NoError(t, err)

	handler.ServeHTTP(wr, req)
	require.Equal(t, http.StatusOK, wr.Code)
}
