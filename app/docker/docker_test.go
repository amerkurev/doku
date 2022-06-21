package docker

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Client(t *testing.T) {
	version := "v1.22"
	addr := "127.0.0.1:3001"
	tmp := t.TempDir()
	mock := NewMockServer(addr, version, tmp, tmp)
	mock.Start(t)

	time.Sleep(10 * time.Millisecond)

	// bad host
	_, err := NewClient(addr, "", version, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to parse docker host")

	// bad certPath
	_, err = NewClient("http://"+addr, t.TempDir(), version, true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not read CA certificate")

	// OK
	ctx := context.Background()
	d, err := NewClient("http://"+addr, "", version, false)
	require.NoError(t, err)

	_, err = d.Ping(ctx)
	assert.NoError(t, err)

	// get docker server version
	ver, err := d.ServerVersion(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "20.10.12", ver.Version)
	assert.Equal(t, "1.41", ver.APIVersion)
	assert.Equal(t, "linux", ver.Os)
	assert.Equal(t, "arm64", ver.Arch)

	// get data usage information.
	du, err := d.DiskUsage(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "alpine:edge", du.Containers[0].Image)
	assert.Equal(t, int64(5349283), du.Containers[0].SizeRootFs)
	assert.Equal(t, int64(0), du.Images[0].Containers)
	assert.Equal(t, int64(899710), du.Volumes[0].UsageData.Size)
	assert.Equal(t, int64(0), du.BuildCache[0].Size)

	// get a list of detailed information about containers
	_, err = d.ContainerJSONList(ctx)
	assert.NoError(t, err)

	// get a stream of events in the docker daemon
	eventCount, err := getEvents(ctx, d)
	assert.True(t, errors.Is(err, io.EOF))
	assert.Equal(t, 8, eventCount)

	mock.Shutdown(t)
}

func getEvents(ctx context.Context, d *Client) (int, error) {
	messages, errs := d.Events(ctx, types.EventsOptions{})
	eventCount := 0
	for {
		select {
		case <-messages:
			eventCount++
		case err := <-errs:
			return eventCount, err
		}
	}
}

func Test_IsSignificantEvent(t *testing.T) {
	tbl := []struct {
		EventType string
		OK        bool
	}{
		{events.BuilderEventType, true},
		{events.ContainerEventType, true},
		{events.DaemonEventType, false},
		{events.ImageEventType, true},
		{events.NetworkEventType, false},
		{events.PluginEventType, false},
		{events.VolumeEventType, true},
		{events.ServiceEventType, true},
		{events.NodeEventType, false},
		{events.SecretEventType, false},
		{events.ConfigEventType, false},
	}

	for _, tt := range tbl {
		assert.Equal(t, tt.OK, IsSignificantEvent(tt.EventType))
	}
}
