package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"testing"
	"time"

	dockerTypes "github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/amerkurev/doku/app/docker"
	"github.com/amerkurev/doku/app/types"
	"github.com/amerkurev/doku/app/util"
)

func Test_Main(t *testing.T) {
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

	dockerVersion := "v1.22"
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	dockerPort := 1000 + rnd.Intn(10000)
	dockerAddr := fmt.Sprintf("127.0.0.1:%d", dockerPort)
	mock := docker.NewMockServer(dockerAddr, dockerVersion, logFile, mountDir)
	mock.Start(t)
	waitForHTTPServerStart(dockerAddr)

	port := 1000 + rnd.Intn(10000)
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	titleHTML := "Doku test"
	os.Args = []string{
		"test",
		"--listen=" + addr,
		"--docker.host=http://" + dockerAddr,
		"--docker.version=" + dockerVersion,
		"--volume=root:/",
		"--log.stdout",
		"--log.level=debug",
		"--ui.home=../web/doku/public", // for index.html
		"--ui.title=" + titleHTML,
	}

	done := make(chan struct{})
	go func() {
		<-done
		err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		require.NoError(t, err)
	}()

	finished := make(chan struct{})
	go func() {
		main()
		close(finished)
	}()

	// defer cleanup because require check below can fail
	defer func() {
		close(done)
		<-finished
		mock.Shutdown(t)
	}()

	waitForHTTPServerStart(addr)
	time.Sleep(time.Second)

	client := http.Client{}
	{
		resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/", port)) // index.html
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)
		b, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Contains(t, string(b), titleHTML)
	}

	{
		resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/v0/version", port))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)
		b, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		ver := types.AppVersion{}
		err = json.Unmarshal(b, &ver)
		require.NoError(t, err)
	}

	{
		resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/v0/disk-usage", port))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)
		b, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		du := util.DiskUsage{}
		err = json.Unmarshal(b, &du)
		require.NoError(t, err)
		assert.Greater(t, du.Total, du.Free)
		assert.GreaterOrEqual(t, du.Free, du.Available)
		assert.GreaterOrEqual(t, du.Percent, 0.)
		assert.LessOrEqual(t, du.Percent, 100.)
		assert.Greater(t, du.Used, uint64(0))
	}

	{
		resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/v0/docker/disk-usage", port))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)
		b, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		du := dockerTypes.DiskUsage{}
		err = json.Unmarshal(b, &du)
		require.NoError(t, err)
	}
}

func waitForHTTPServerStart(addr string) {
	// wait for up to 10 seconds for server to start before returning it
	client := http.Client{Timeout: time.Second}
	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond * 100)
		if resp, err := client.Get("http://" + addr + "/version"); err == nil {
			_ = resp.Body.Close()
			return
		}
	}
}

func Test_configureLogging(t *testing.T) {

	tbl := []struct {
		opt   string
		level log.Level
	}{
		{"debug", log.DebugLevel},
		{"info", log.InfoLevel},
		{"warning", log.WarnLevel},
		{"error", log.ErrorLevel},
	}

	for i, tt := range tbl {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			configureLogging(tt.opt, true)
			assert.Equal(t, tt.level, log.GetLevel())
		})
	}
}

func Test_listenAddress(t *testing.T) {
	s := listenAddress("")
	assert.Equal(t, "127.0.0.1:9090", s)

	t.Setenv("DOKU_IN_DOCKER", "1")
	s = listenAddress("")
	assert.Equal(t, "0.0.0.0:9090", s)

	t.Setenv("DOKU_IN_DOCKER", "true")
	s = listenAddress("")
	assert.Equal(t, "0.0.0.0:9090", s)

	addr := "127.0.0.1:80"
	s = listenAddress(addr)
	assert.Equal(t, addr, s)
}

func Test_makeBasicAuth(t *testing.T) {
	pf := `test:$2y$05$zMxDmK65SjcH2vJQNopVSO/nE8ngVLx65RoETyHpez7yTS/8CLEiW
		test2:$2y$05$TLQqHh6VT4JxysdKGPOlJeSkkMsv.Ku/G45i7ssIm80XuouCrES12
		bad bad`

	fh, err := os.CreateTemp(t.TempDir(), "doku_auth_*")
	require.NoError(t, err)
	defer fh.Close()

	n, err := fh.WriteString(pf)
	require.NoError(t, err)
	require.Equal(t, len(pf), n)

	res, err := makeBasicAuth(fh.Name())
	require.NoError(t, err)
	assert.Equal(t, 3, len(res))
	assert.Equal(t, []string{"test:$2y$05$zMxDmK65SjcH2vJQNopVSO/nE8ngVLx65RoETyHpez7yTS/8CLEiW", "test2:$2y$05$TLQqHh6VT4JxysdKGPOlJeSkkMsv.Ku/G45i7ssIm80XuouCrES12", "bad bad"}, res)

	// no such file or directory
	_, err = makeBasicAuth("incorrect-path")
	var perr *fs.PathError
	require.ErrorAs(t, err, &perr)
}

func Test_parseVolumes(t *testing.T) {
	tbl := []struct {
		input []string
		vols  []types.HostVolume
		err   error
	}{
		{[]string{"data volume:/data"}, []types.HostVolume{{Name: "data volume", Path: "/data"}}, nil},
		{[]string{"data volume:/data", "blah:/"}, []types.HostVolume{{Name: "data volume", Path: "/data"}, {Name: "blah", Path: "/"}}, nil},
		{[]string{"/data"}, []types.HostVolume{}, errors.New("invalid volume format, should be <name>:<path>")},
	}

	for i, tt := range tbl {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			vols, err := parseVolumes(tt.input)
			if tt.err != nil {
				require.EqualError(t, err, tt.err.Error())
				return
			}
			assert.Equal(t, tt.vols, vols)
		})
	}
}
