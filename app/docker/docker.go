package docker

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"

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

// Client defines Docker client.
type Client struct {
	*docker.Client
}

// Events returns a stream of events in the Docker daemon.
func (c *Client) Events(ctx context.Context) (<-chan events.Message, <-chan error) {
	return c.Client.Events(ctx, types.EventsOptions{})
}

// Info returns information about the Docker server.
func (c *Client) Info(ctx context.Context) (types.Info, error) {
	return c.Client.Info(ctx)
}

// DiskUsage requests the current data usage from the Docker daemon.
func (c *Client) DiskUsage(ctx context.Context) (types.DiskUsage, error) {
	return c.Client.DiskUsage(ctx)
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

// BindMounts returns all files or directories on the host machine that mounted into containers.
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
