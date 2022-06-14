package docker

import (
	"context"
	"net/http"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/tlsconfig"
	log "github.com/sirupsen/logrus"
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
type Client interface {
	Events(context.Context) (<-chan events.Message, <-chan error)
	Info(context.Context) *Info
}

type client struct {
	*docker.Client
}

// NewClient creates a new Docker client.
func NewClient(host, certPath, version string, verify bool) (Client, error) {

	cli, err := docker.NewClientWithOpts(func(c *docker.Client) error {
		return setOpts(c, host, certPath, version, verify)
	})

	if err != nil {
		return nil, err
	}

	c := &client{
		Client: cli,
	}
	return c, nil
}

// Events returns a stream of events in the Docker daemon.
func (c *client) Events(ctx context.Context) (<-chan events.Message, <-chan error) {
	return c.Client.Events(ctx, types.EventsOptions{})
}

func (c *client) Info(ctx context.Context) *Info {
	defer elapsed("yet another poll execution is done")()

	info, err := c.Client.Info(ctx)
	if err != nil {
		log.WithField("err", err).Error("docker info request error")
	}

	diskUsage, err := c.Client.DiskUsage(ctx)
	if err != nil {
		log.WithField("err", err).Error("docker disk usage request error")
	}

	return &Info{
		Info:      info,
		DiskUsage: diskUsage,
		Timestamp: time.Now().UnixMicro(),
	}
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

func elapsed(what string) func() {
	start := time.Now()
	return func() {
		log.WithField("took", time.Since(start)).Debug(what)
	}
}
