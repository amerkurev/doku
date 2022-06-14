package poller

import (
	"context"
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
	messages, errs := d.Events(ctx)
	numMessages := 0 // count Docker daemon events.

	go func() {
		// Run it immediately on start
		_ = d.Info(ctx)

		// Run poll with interval while context is not cancel
		for {
			select {
			case m := <-messages:
				if docker.IsSignificantEvent(m.Type) {
					numMessages += 1
				}
			case err = <-errs:
				if err != nil {
					log.WithField("err", err).Error("docker event listener error")

					// Reconnect to Docker daemon
					select {
					case <-time.After(pollingLongInterval):
						messages, errs = d.Events(ctx)
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
					_ = d.Info(ctx)
				}
			case <-time.After(pollingLongInterval):
				_ = d.Info(ctx)
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
