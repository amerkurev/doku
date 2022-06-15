package poller

import (
	"context"
	"encoding/json"
	"github.com/amerkurev/doku/app/store"
	"time"

	"github.com/amerkurev/doku/app/docker"
	log "github.com/sirupsen/logrus"
)

const (
	pollingShortInterval = time.Second
	pollingLongInterval  = time.Minute
)

// Poller represents an interface of Docker polling control.
type Poller interface {
	Stop()
}

type poller struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// New creates a poller and starts a goroutine to poll the Docker daemon.
func New(host, certPath, version string, verify bool) (Poller, error) {
	d, err := docker.NewClient(host, certPath, version, verify)
	if err != nil {
		return nil, err
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	messages, errs := d.DockerEvents(ctx)
	numMessages := 0 // count Docker daemon events.

	go func() {
		// Run it immediately on start
		save(d.DockerInfo(ctx))

		// Run poll with interval while context is not cancel
		for {
			select {
			case m := <-messages:
				if docker.IsSignificantEvent(m.Type) {
					numMessages++
				}
			case err = <-errs:
				if err != nil {
					log.WithField("err", err).Error("docker event listener error")

					// Reconnect to Docker daemon
					select {
					case <-time.After(pollingLongInterval):
						messages, errs = d.DockerEvents(ctx)
					case <-ctx.Done():
						return
					}
				}
			case <-ctx.Done():
				return
			case <-time.After(pollingShortInterval):
				// Execute poll only if received Docker daemon events.
				if numMessages > 0 {
					numMessages = 0
					save(d.DockerInfo(ctx))
				}
			case <-time.After(pollingLongInterval):
				save(d.DockerInfo(ctx))
			}
		}
	}()

	p := &poller{
		ctx:        ctx,
		cancelFunc: cancelFunc,
	}
	return p, nil
}

// Stop stops Docker daemon polling.
func (p *poller) Stop() {
	p.cancelFunc()
}

func save(d *docker.Info) {
	if d != nil {
		s := store.Get()
		s.Set("latestPolling", d)

		b, err := json.Marshal(d.Info)
		if err != nil {
			log.WithField("err", err).Error("docker info serialization error")
		}
		s.Set("json.dockerInfo", b)

		b, err = json.Marshal(d.DiskUsage)
		if err != nil {
			log.WithField("err", err).Error("docker disk usage serialization error")
		}
		s.Set("json.dockerDiskUsage", b)

		// wake up those who are waiting
		s.Notify()
	}
}
