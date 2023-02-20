package middleware

import (
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SinglePageApplication(t *testing.T) {
	// wrong template
	fhErr, err := os.CreateTemp(t.TempDir(), "doku_spa_*")
	require.NoError(t, err)
	defer fhErr.Close()
	_, err = fhErr.WriteString("{{ .DoesExist }}")
	require.NoError(t, err)

	tbl := []struct {
		indexHTML string
		title     string
		header    string
		code      int
	}{
		{indexHTML: "../../../web/doku/public/index.html", title: "YgsGWsIASy8sUnDF", header: "YsLExc8bsrviguGv", code: 200},
		{indexHTML: fhErr.Name(), title: "", header: "", code: 200},
		{indexHTML: "index.html", title: "YgsGWsIASy8sUnDF", header: "YsLExc8bsrviguGv", code: 404},
	}

	for i, tt := range tbl {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			wr := httptest.NewRecorder()
			handler := SinglePageApplication(tt.indexHTML, tt.title, tt.header)

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
