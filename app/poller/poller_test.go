package poller

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/amerkurev/doku/app/docker"
	"github.com/amerkurev/doku/app/store"
	"github.com/amerkurev/doku/app/types"
)

func Test_Run(t *testing.T) {
	var err error
	tmp := t.TempDir()

	// prepare log dir
	logDir := path.Join(tmp, "logs")
	err = os.Mkdir(logDir, 0777)
	require.NoError(t, err)
	logPath := path.Join(logDir, "container-log.json")
	err = os.WriteFile(logPath, []byte("some content"), 0644)
	require.NoError(t, err)

	// prepare mount dir
	mountDir := path.Join(tmp, "mount")
	require.NoError(t, err)
	err = os.Mkdir(mountDir, 0777)
	require.NoError(t, err)
	require.NoError(t, err)
	err = os.WriteFile(path.Join(mountDir, "payload.txt"), []byte("some content"), 0644)
	require.NoError(t, err)

	// options
	version := "v1.22"
	addr := "127.0.0.1:3000"
	mock := docker.NewMockServer(addr, version, logPath, mountDir)
	mock.Start(t)

	volumes := []types.HostVolume{
		{Name: "root", Path: "/"},
	}

	time.Sleep(10 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	d, err := docker.NewClient("http://"+addr, "", version, false)
	require.NoError(t, err)

	err = store.Initialize()
	require.NoError(t, err)

	Run(ctx, d, volumes)
	time.Sleep(3 * time.Second)

	cancel()
	mock.Shutdown(t)

	v, ok := store.Get("dockerVersion")
	assert.True(t, ok)
	assert.Equal(t, 825, len(v.([]byte)))
	v, ok = store.Get("dockerDiskUsage")
	assert.True(t, ok)
	assert.Equal(t, 1956, len(v.([]byte)))
	v, ok = store.Get("dockerLogInfo")
	assert.True(t, ok)
	assert.Equal(t, 330, len(v.([]byte)))
	v, ok = store.Get("dockerMountsBind")
	assert.True(t, ok)
	assert.Equal(t, 246, len(v.([]byte)))
	v, ok = store.Get("sizeCalcProgress")
	assert.True(t, ok)
	assert.Equal(t, 72, len(v.([]byte)))
}

func Test_RunNoSuchFileOrDir(t *testing.T) {
	var err error
	tmp := t.TempDir()

	// prepare log dir
	logDir := path.Join(tmp, "logs")
	logPath := path.Join(logDir, "container-log.json")

	// prepare mount dir
	mountDir := path.Join(tmp, "mount")

	// options
	version := "v1.22"
	addr := "127.0.0.1:3000"
	mock := docker.NewMockServer(addr, version, logPath, mountDir)
	mock.Start(t)

	volumes := []types.HostVolume{
		{Name: "root", Path: "/hostroot"},
	}

	time.Sleep(10 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	d, err := docker.NewClient("http://"+addr, "", version, false)
	require.NoError(t, err)

	err = store.Initialize()
	require.NoError(t, err)

	Run(ctx, d, volumes)
	time.Sleep(3 * time.Second)

	cancel()
	mock.Shutdown(t)
}
