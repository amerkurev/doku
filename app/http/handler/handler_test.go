package handler

import (
	"context"
	"fmt"
	"math/rand"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/amerkurev/doku/app/docker"
	"github.com/amerkurev/doku/app/store"
	"github.com/amerkurev/doku/app/types"
)

var revision = "unknown"

func Test_Handler(t *testing.T) {
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

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	port := 10000 + rnd.Intn(10000)
	dockerMockAddr := fmt.Sprintf("127.0.0.1:%d", port)

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

	rr := httptest.NewRecorder()
	internalServerError(rr, err, "reason")
	assert.Equal(t, 500, rr.Result().StatusCode)

	rr = httptest.NewRecorder()
	Version(ctx)(rr, nil)
	assert.Equal(t, 200, rr.Result().StatusCode)
	assert.Equal(t, fmt.Sprintf("{\"version\":\"%s\"}", revision), rr.Body.String())

	rr = httptest.NewRecorder()
	DockerVersion(ctx, d)(rr, nil)
	assert.Equal(t, 200, rr.Result().StatusCode)
	assert.Greater(t, rr.Body.Len(), 0)

	rr = httptest.NewRecorder()
	DiskUsage(ctx)(rr, nil)
	assert.Equal(t, 200, rr.Result().StatusCode)
	assert.Greater(t, rr.Body.Len(), 0)

	rr = httptest.NewRecorder()
	DockerDiskUsage(ctx, d)(rr, nil)
	assert.Equal(t, 200, rr.Result().StatusCode)
	assert.Greater(t, rr.Body.Len(), 0)

	rr = httptest.NewRecorder()
	DockerContainerList(ctx, d)(rr, nil)
	assert.Equal(t, 200, rr.Result().StatusCode)
	assert.Greater(t, rr.Body.Len(), 0)

	rr = httptest.NewRecorder()
	DockerLogSize(ctx, d)(rr, nil)
	assert.Equal(t, 200, rr.Result().StatusCode)
	assert.Greater(t, rr.Body.Len(), 0)

	rr = httptest.NewRecorder()
	store.Set("dockerBindMounts", []byte("{}"))
	DockerBindMounts()(rr, nil)
	assert.Equal(t, 200, rr.Result().StatusCode)
	assert.Equal(t, "{}", rr.Body.String())

	mock.Shutdown(t)

	// failed
	rr = httptest.NewRecorder()
	DockerVersion(ctx, d)(rr, nil)
	assert.Equal(t, 500, rr.Result().StatusCode)

	rr = httptest.NewRecorder()
	DockerDiskUsage(ctx, d)(rr, nil)
	assert.Equal(t, 500, rr.Result().StatusCode)

	rr = httptest.NewRecorder()
	DockerContainerList(ctx, d)(rr, nil)
	assert.Equal(t, 500, rr.Result().StatusCode)

	rr = httptest.NewRecorder()
	DockerLogSize(ctx, d)(rr, nil)
	assert.Equal(t, 500, rr.Result().StatusCode)
}

func Test_FailedToGetLogFile(t *testing.T) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	port := 10000 + rnd.Intn(10000)
	dockerMockAddr := fmt.Sprintf("127.0.0.1:%d", port)

	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, types.CtxKeyRevision, revision)
	ctx = context.WithValue(ctx, types.CtxKeyVolumes, []types.HostVolume{{Name: "root", Path: "/"}})
	defer cancel()

	// Docker mock
	version := "v1.22"
	mock := docker.NewMockServer(dockerMockAddr, version, "incorrect-path", "incorrect-path")
	mock.Start(t)
	d, err := docker.NewClient(ctx, "http://"+dockerMockAddr, "", version, false)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	DockerLogSize(ctx, d)(rr, nil)
	assert.Equal(t, 200, rr.Result().StatusCode)
	assert.Greater(t, rr.Body.Len(), 0)
}
