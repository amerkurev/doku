package handler

import (
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SinglePageApplication(t *testing.T) {
	tbl := []struct {
		staticFolder string
		title        string
		header       string
		code         int
	}{
		{staticFolder: "../../../frontend/static", title: "YgsGWsIASy8sUnDF", header: "YsLExc8bsrviguGv", code: 200},
		{staticFolder: "../../../testdata", title: "", header: "", code: 200},
		{staticFolder: "", title: "YgsGWsIASy8sUnDF", header: "YsLExc8bsrviguGv", code: 404},
	}

	for i, tt := range tbl {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			wr := httptest.NewRecorder()
			handler := SinglePageApplication(tt.staticFolder, tt.title, tt.header)

			req, err := http.NewRequest("GET", "http://example.com", http.NoBody)
			require.NoError(t, err)

			handler.ServeHTTP(wr, req)
			require.Equal(t, tt.code, wr.Code)

			if wr.Code == http.StatusOK {
				b, err := io.ReadAll(wr.Body)
				assert.NoError(t, err)
				assert.Contains(t, string(b), tt.title)
				assert.Contains(t, string(b), tt.header)
			}
		})
	}
}
