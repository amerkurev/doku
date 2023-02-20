package http

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/amerkurev/doku/app/docker"
	"github.com/amerkurev/doku/app/types"
)

var revision = "unknown"

func Test_Server_Run(t *testing.T) {
	var err error
	s := "some content"

	// prepare log file
	fh, err := os.CreateTemp(t.TempDir(), "doku_log_*")
	require.NoError(t, err)
	defer fh.Close()
	n, err := fh.WriteString(s)
	require.NoError(t, err)
	require.Equal(t, len(s), n)
	logFile := fh.Name()

	// prepare mount dir
	mountDir := t.TempDir()
	fh, err = os.CreateTemp(mountDir, "doku_mount_*")
	require.NoError(t, err)
	defer fh.Close()
	n, err = fh.WriteString(s)
	require.NoError(t, err)
	require.Equal(t, len(s), n)

	err = os.Setenv("ENVIRONMENT", "dev") // CORS
	require.NoError(t, err)

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	port := 1000 + rnd.Intn(10000)
	dockerMockAddr := fmt.Sprintf("127.0.0.1:%d", port)

	port = 1000 + rnd.Intn(10000)
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	allowed := []string{
		"user:$2y$10$ViWb72wAg4nEKdHh.9p2yeEKLSN0EBkbZ0Mf0bqNHZmItsQt6K8he",
	}

	httpServer := &Server{
		Address: addr,
		Timeouts: Timeouts{
			Read:     5 * time.Second,
			Write:    60 * time.Second,
			Idle:     60 * time.Second,
			Shutdown: 5 * time.Second,
		},
		BasicAuthEnabled: true,
		BasicAuthAllowed: allowed,
		StaticFolder:     "../../web/doku/public", // for index.html
	}

	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, types.CtxKeyRevision, revision)
	ctx = context.WithValue(ctx, types.CtxKeyVolumes, []types.HostVolume{{Name: "root", Path: "/"}})
	defer cancel()

	// Docker mock
	version := "v1.22"
	mock := docker.NewMockServer(dockerMockAddr, version, logFile, mountDir)
	mock.Start(t)
	d, err := docker.NewClient(ctx, "http://"+dockerMockAddr, "", version, false)
	require.NoError(t, err)

	done := make(chan struct{})
	go func() {
		err := httpServer.Run(ctx, d)
		assert.NoError(t, err)
		done <- struct{}{}
	}()

	tbl := []struct {
		endpoint string
	}{
		{"/"},                      // 0
		{"/favicon.ico"},           // 1
		{"/manifest.json"},         // 2
		{"/v0/version"},            // 3
		{"/v0/disk-usage"},         // 4
		{"/v0/docker/version"},     // 5
		{"/v0/docker/containers"},  // 6
		{"/v0/docker/disk-usage"},  // 7
		{"/v0/docker/log-size"},    // 8
		{"/v0/docker/bind-mounts"}, // 9
	}

	time.Sleep(10 * time.Millisecond)

	client := http.Client{}

	for i, tt := range tbl {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			req, err := http.NewRequest("GET", "http://127.0.0.1:"+strconv.Itoa(port)+tt.endpoint, http.NoBody)
			require.NoError(t, err)
			req.SetBasicAuth("user", "1111")
			resp, err := client.Do(req)
			require.NoError(t, err)
			err = resp.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}

	cancel()
	<-done
	mock.Shutdown(t)
}

func Test_Server_RunFailed(t *testing.T) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	port := 1000 + rnd.Intn(10000)
	dockerMockAddr := fmt.Sprintf("127.0.0.1:%d", port)

	port = 1_000_000
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	httpServer := &Server{
		Address: addr,
		Timeouts: Timeouts{
			Read:     5 * time.Second,
			Write:    60 * time.Second,
			Idle:     60 * time.Second,
			Shutdown: 5 * time.Second,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	ctx = context.WithValue(ctx, types.CtxKeyRevision, revision)
	ctx = context.WithValue(ctx, types.CtxKeyVolumes, []types.HostVolume{{Name: "root", Path: "/"}})
	defer cancel()

	// Docker mock
	version := "v1.22"
	mock := docker.NewMockServer(dockerMockAddr, version, "", "")
	mock.Start(t)
	d, err := docker.NewClient(ctx, "http://"+addr, "", version, false)
	require.NoError(t, err)

	done := make(chan struct{})
	go func() {
		err := httpServer.Run(ctx, d)
		assert.Error(t, err)
		assert.EqualError(t, errors.New("http server failed: listen tcp: address 1000000: invalid port"), err.Error())
		done <- struct{}{}
	}()

	<-done
	mock.Shutdown(t)
}
