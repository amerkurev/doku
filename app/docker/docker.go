package docker

import (
	"context"
	"net/http"
	"path/filepath"

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

// ContainerJSONList returns the list of the container information.
func (c *Client) ContainerJSONList(ctx context.Context) ([]*types.ContainerJSON, error) {
	containers, err := c.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}

	res := make([]*types.ContainerJSON, 0, len(containers))

	for _, cont := range containers {
		ci, _, err := c.ContainerInspectWithRaw(ctx, cont.ID, true)
		if err != nil {
			return nil, err
		}
		res = append(res, &ci)
	}
	return res, nil
}

// NewClient creates a new Docker client.
func NewClient(ctx context.Context, host, certPath, version string, verify bool) (*Client, error) {
	cli, err := docker.NewClientWithOpts(func(c *docker.Client) error {
		return setOpts(c, host, certPath, version, verify)
	})

	if err != nil {
		return nil, err
	}

	cli.NegotiateAPIVersion(ctx)
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
