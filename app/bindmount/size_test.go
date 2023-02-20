package bindmount

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/amerkurev/doku/app/docker"
	"github.com/amerkurev/doku/app/store"
	"github.com/amerkurev/doku/app/types"
)

func Test_Run(t *testing.T) {
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

	// options
	version := "v1.22"
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	port := 10000 + rnd.Intn(10000)
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	mock := docker.NewMockServer(addr, version, logFile, mountDir)
	mock.Start(t)
	time.Sleep(10 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	d, err := docker.NewClient(ctx, "http://"+addr, "", version, false)
	require.NoError(t, err)

	require.NoError(t, err)

	volumes := []types.HostVolume{{Name: "root", Path: "/"}}

	CalcSize(ctx, d, volumes)
	time.Sleep(2 * time.Second)

	cancel()
	mock.Shutdown(t)

	v, ok := store.Get("dockerBindMounts")
	assert.True(t, ok)
	bindMounts := struct {
		BindMounts []*types.BindMountInfo
		TotalSize  int64
	}{}
	err = json.Unmarshal(v.([]byte), &bindMounts)
	require.NoError(t, err)
}

func Test_Run_NoSuchFileOrDir(t *testing.T) {
	// options
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	port := 10000 + rnd.Intn(10000)
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	version := "v1.22"
	mock := docker.NewMockServer(addr, version, "incorrect-path", "incorrect-path")
	mock.Start(t)
	time.Sleep(10 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	d, err := docker.NewClient(ctx, "http://"+addr, "", version, false)
	require.NoError(t, err)

	require.NoError(t, err)

	volumes := []types.HostVolume{{Name: "root", Path: "/"}}

	CalcSize(ctx, d, volumes)
	time.Sleep(time.Second)

	cancel()
	mock.Shutdown(t)
}

func Test_Run_Failed(t *testing.T) {
	// options
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	port := 10000 + rnd.Intn(10000)
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	mock := docker.NewMockServer(addr, "", "", "")
	mock.Start(t)
	time.Sleep(10 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	d, err := docker.NewClient(ctx, "http://"+addr, "", "", false)
	require.NoError(t, err)

	require.NoError(t, err)

	volumes := []types.HostVolume{{Name: "root", Path: "/"}}

	CalcSize(ctx, d, volumes)

	cancel()
	mock.Shutdown(t)
}

func Test_contains(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	assert.True(t, contains[int](5, numbers))
	assert.False(t, contains[int](11, numbers))

	str := []string{"a", "b", "c"}
	assert.True(t, contains[string]("a", str))
	assert.False(t, contains[string]("d", str))
}

func Test_bindMountsTotalSize(t *testing.T) {

	tbl := []struct {
		Mounts            []*types.BindMountInfo
		ExpectedTotalSize int64
	}{
		{[]*types.BindMountInfo{}, 0},

		{[]*types.BindMountInfo{
			{Path: "/a", Size: 1000},
		}, 1000},

		{[]*types.BindMountInfo{
			{Path: "/a", Size: 0},
			{Path: "/a/b", Size: 10},
		}, 10},

		{[]*types.BindMountInfo{
			{Path: "/a", Size: 100},
			{Path: "/a/b", Size: 40},
			{Path: "/a/b/c", Size: 10},
		}, 100},

		{[]*types.BindMountInfo{
			{Path: "/a/b", Size: 100},
			{Path: "/c/d", Size: 40},
			{Path: "/e/a/b", Size: 10},
		}, 150},

		{[]*types.BindMountInfo{
			{Path: "/a/b/c", Size: 100},
			{Path: "/a/b/e", Size: 40},
			{Path: "/a/b/e/d", Size: 10},
		}, 140},

		{[]*types.BindMountInfo{
			{Path: "/a/b/c", Size: 100},
			{Path: "/a/b", Size: 40},
			{Path: "/a", Size: 40},
			{Path: "/d", Size: 10},
			{Path: "/d/e/f", Size: 10},
			{Path: "/d/f", Size: 10},
			{Path: "/x", Size: 10},
			{Path: "/y", Size: 10},
			{Path: "/z", Size: 10},
		}, 80},
	}

	for i, tt := range tbl {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			actual := bindMountsTotalSize(tt.Mounts)
			assert.Equal(t, tt.ExpectedTotalSize, actual)
		})
	}
}
