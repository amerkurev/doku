package docker

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/tlsconfig"
)

var significantDockerEvents = [...]string{
	events.BuilderEventType,
	events.ContainerEventType,
	events.ImageEventType,
	events.VolumeEventType,
	events.ServiceEventType,
}

// Info contains all information that retrieves from the Docker daemon.
type Info struct {
	types.Info
	types.DiskUsage
	Timestamp int64
}

// Client defines Docker client.
type Client struct {
	*docker.Client
}

// NewClient creates a new Docker client.
func NewClient(host, certPath, version string, verify bool) (*Client, error) {

	cli, err := docker.NewClientWithOpts(func(c *docker.Client) error {
		return setOpts(c, host, certPath, version, verify)
	})

	if err != nil {
		return nil, err
	}

	return &Client{cli}, nil
}

// DockerEvents returns a stream of events in the Docker daemon.
func (c *Client) DockerEvents(ctx context.Context) (<-chan events.Message, <-chan error) {
	return c.Client.Events(ctx, types.EventsOptions{})
}

// DockerInfo returns a piece of information about disk usage and others.
func (c *Client) DockerInfo(ctx context.Context) (*Info, error) {
	info, err := c.Client.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("info request: %w", err)
	}

	diskUsage, err := c.Client.DiskUsage(ctx)
	if err != nil {
		return nil, fmt.Errorf("disk usage request: %w", err)
	}

	r := &Info{
		Info:      info,
		DiskUsage: diskUsage,
		Timestamp: time.Now().UnixMicro(),
	}
	return r, nil
}

func setOpts(c *docker.Client, host, certPath, version string, verify bool) error {
	if certPath != "" {
		options := tlsconfig.Options{
			CAFile:             filepath.Join(certPath, "ca.pem"),
			CertFile:           filepath.Join(certPath, "cert.pem"),
			KeyFile:            filepath.Join(certPath, "key.pem"),
			InsecureSkipVerify: !verify,
		}
		tlsc, err := tlsconfig.Client(options)
		if err != nil {
			return err
		}

		httpClient := &http.Client{
			Transport:     &http.Transport{TLSClientConfig: tlsc},
			CheckRedirect: docker.CheckRedirect,
		}

		if err := docker.WithHTTPClient(httpClient)(c); err != nil {
			return err
		}
	}

	if host != "" {
		if err := docker.WithHost(host)(c); err != nil {
			return err
		}
	}

	if version != "" {
		if err := docker.WithVersion(version)(c); err != nil {
			return err
		}
	}
	return nil
}

// IsSignificantEvent says that event is relative to disk usage.
func IsSignificantEvent(e string) bool {
	for _, event := range significantDockerEvents {
		if e == event {
			return true
		}
	}
	return false
}

func BindMounts(containers []*types.Container) []string {
	res := make([]string, 0, len(containers))
	for _, c := range containers {
		for _, m := range c.Mounts {
			if m.Type == "bind" {
				res = append(res, strings.TrimPrefix(m.Source, "/host_mnt"))
			}
		}
	}
	return res
}
