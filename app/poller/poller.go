package poller

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/amerkurev/doku/app/docker"
	"github.com/amerkurev/doku/app/store"
	"github.com/amerkurev/doku/app/types"
	"github.com/amerkurev/doku/app/util"
	log "github.com/sirupsen/logrus"
)

const (
	pollingShortInterval = time.Second
	pollingLongInterval  = time.Minute
)

// Run starts a goroutine to poll the Docker daemon.
func Run(ctx context.Context, d *docker.Client, volumes []types.HostVolume) {
	messages, errs := d.Events(ctx)
	numMessages := 0 // count of Docker daemon events.

	// calculate the size of directories that mounted into the containers on the host machine
	dirSizeCalculator(ctx, d, volumes)

	go func() {
		// run it immediately on start
		poll(ctx, d)

		// run poll with interval while context is not cancel
		for {
			select {
			case m := <-messages:
				if docker.IsSignificantEvent(m.Type) {
					numMessages++
				}
			case err := <-errs:
				if err != nil {
					log.WithField("err", err).Error("failed to listen to docker events")

					// reconnect to the Docker daemon
					select {
					case <-time.After(pollingLongInterval):
						messages, errs = d.Events(ctx)
					case <-ctx.Done():
						log.Info("gracefully poller shutdown")
						return
					}
				}
			case <-ctx.Done():
				log.Info("gracefully poller shutdown")
				return
			case <-time.After(pollingShortInterval):
				// execute poll only if was happened Docker daemon events
				if numMessages > 0 {
					numMessages = 0
					poll(ctx, d)
				}
			case <-time.After(pollingLongInterval):
				// forced poll every minute
				poll(ctx, d)
			}
		}
	}()
}

func poll(ctx context.Context, d *docker.Client) {
	defer util.PrintExecTime("yet another poll execution is done")()
	defer store.NotifyAll() // wake up those who are waiting.

	if err := dockerInfo(ctx, d); err != nil {
		log.WithField("err", err).Error("failed to get information about the docker server")
	}

	if err := dockerDiskUsage(ctx, d); err != nil {
		log.WithField("err", err).Error("failed to request the current data usage from the docker daemon")
	}
}

func dockerInfo(ctx context.Context, d *docker.Client) error {
	r, err := d.Info(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute request")
	}

	b, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("failed to encode as JSON")
	}

	store.Set("dockerInfo", b)
	return nil
}

func dockerDiskUsage(ctx context.Context, d *docker.Client) error {
	r, err := d.DiskUsage(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute request")
	}

	b, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("failed to encode as JSON")
	}

	store.Set("dockerDiskUsage", b)
	return nil
}
