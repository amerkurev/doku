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

// Run starts a goroutine to poll the Docker daemon.
func Run(ctx context.Context, d *docker.Client) {
	messages, errs := d.DockerEvents(ctx)
	numMessages := 0 // count Docker daemon events.

	go func() {
		// Run it immediately on start
		poll(ctx, d)

		// Run poll with interval while context is not cancel
		for {
			select {
			case m := <-messages:
				if docker.IsSignificantEvent(m.Type) {
					numMessages++
				}
			case err := <-errs:
				if err != nil {
					log.WithField("err", err).Error("failed to listen to docker events")

					// Reconnect to Docker daemon
					select {
					case <-time.After(pollingLongInterval):
						messages, errs = d.DockerEvents(ctx)
					case <-ctx.Done():
						log.Info("gracefully poller shutdown")
						return
					}
				}
			case <-ctx.Done():
				log.Info("gracefully poller shutdown")
				return
			case <-time.After(pollingShortInterval):
				// Execute poll only if received Docker daemon events.
				if numMessages > 0 {
					numMessages = 0
					poll(ctx, d)
				}
			case <-time.After(pollingLongInterval):
				poll(ctx, d)
			}
		}
	}()
}

func poll(ctx context.Context, d *docker.Client) {
	defer elapsed("yet another poll execution is done")()

	r, err := d.DockerInfo(ctx)
	if err != nil {
		log.WithField("err", err).Error("failed to docker request")
		return
	}

	s := store.Get()
	s.Set("latestPolling", r)

	b, err := json.Marshal(r.Info)
	if err != nil {
		log.WithField("err", err).Error("failed to serialize docker info")
		return
	}
	s.Set("json.dockerInfo", b)

	b, err = json.Marshal(r.DiskUsage)
	if err != nil {
		log.WithField("err", err).Error("failed to serialize docker disk usage")
		return
	}
	s.Set("json.dockerDiskUsage", b)

	// wake up those who are waiting
	s.NotifyAll()
}

func elapsed(what string) func() {
	start := time.Now()
	return func() {
		log.WithField("took", time.Since(start)).Debug(what)
	}
}
