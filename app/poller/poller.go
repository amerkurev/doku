package poller

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	dockerTypes "github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"

	"github.com/amerkurev/doku/app/docker"
	"github.com/amerkurev/doku/app/store"
	"github.com/amerkurev/doku/app/types"
	"github.com/amerkurev/doku/app/util"
)

const pollingInterval = time.Minute

var logPathErrors map[string]bool // to prevent the output of the same errors in the log

// Run starts a goroutine to poll the Docker daemon.
func Run(ctx context.Context, d *docker.Client, volumes []types.HostVolume) {
	messages, errs := d.Events(ctx, dockerTypes.EventsOptions{})
	numMessages := 0 // count of Docker daemon events.
	logPathErrors = make(map[string]bool)

	// calculate the size of directories that mounted into containers (bind type)
	bindMountsSize(ctx, d, volumes)

	go func() {
		// run it immediately on start
		poll(ctx, d, volumes)
		lastPoll := time.Now()

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
					case <-time.After(pollingInterval):
						messages, errs = d.Events(ctx, dockerTypes.EventsOptions{})
					case <-ctx.Done():
						log.Info("gracefully poller shutdown")
						return
					}
				}
			case <-ctx.Done():
				log.Info("gracefully poller shutdown")
				return
			case <-time.After(50 * time.Millisecond):
				// execute poll only if was happened a few Docker daemon events
				if numMessages > 0 && time.Since(lastPoll) > time.Second {
					numMessages = 0
					poll(ctx, d, volumes)
					lastPoll = time.Now()
				}

				// forced poll in a minute after the last poll
				if time.Since(lastPoll) > pollingInterval {
					poll(ctx, d, volumes)
					lastPoll = time.Now()
				}
			}
		}
	}()
}

func poll(ctx context.Context, d *docker.Client, volumes []types.HostVolume) {
	defer util.PrintExecTime("poll execution progress")()
	defer store.NotifyAll() // wake up those who are waiting.

	if err := dockerVersion(ctx, d); err != nil {
		log.WithField("err", err).Error("poll: docker version")
	}

	if err := dockerContainerList(ctx, d); err != nil {
		log.WithField("err", err).Error("poll: docker container list")
	}

	if err := dockerDiskUsage(ctx, d); err != nil {
		log.WithField("err", err).Error("poll: docker disk usage")
	}

	if err := dockerLogSize(ctx, d, volumes); err != nil {
		log.WithField("err", err).Error("poll: docker log size")
	}
}

func dockerVersion(ctx context.Context, d *docker.Client) error {
	res, err := d.ServerVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	// hide sensitive (and not only) data
	res.Components = nil

	b, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("failed to encode as JSON: %w", err)
	}

	store.Set("dockerVersion", b)
	return nil
}

func dockerContainerList(ctx context.Context, d *docker.Client) error {
	res, err := d.ContainerJSONList(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	var totalSize int64
	for _, c := range res {
		// hide sensitive (and not only) data
		c.HostConfig = nil
		c.GraphDriver.Data = nil
		c.GraphDriver.Name = ""
		c.NetworkSettings = nil
		c.Image = c.Config.Image // Image ID -> Image Tag
		c.Config = nil
		if c.SizeRw != nil {
			totalSize += *c.SizeRw
		}
	}

	b, err := json.Marshal(struct {
		Containers []*dockerTypes.ContainerJSON
		TotalSize  int64
	}{
		Containers: res,
		TotalSize:  totalSize,
	})

	if err != nil {
		return fmt.Errorf("failed to encode as JSON: %w", err)
	}

	store.Set("dockerContainerList", b)
	return nil
}

func dockerDiskUsage(ctx context.Context, d *docker.Client) error {
	res, err := d.DiskUsage(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	// hide sensitive (and not only) data
	res.Containers = nil

	b, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("failed to encode as JSON: %w", err)
	}

	store.Set("dockerDiskUsage", b)
	return nil
}

func dockerLogSize(ctx context.Context, d *docker.Client, volumes []types.HostVolume) error {
	containers, err := d.ContainerJSONList(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	logs := make([]*types.LogFileInfo, 0)
	var totalSize int64

	for _, cont := range containers {
		l, errSize := logFileSize(cont, volumes) // get size of container log file
		if errSize != nil {
			if _, ok := logPathErrors[cont.LogPath]; !ok {
				// to prevent the output of the same errors in the log
				logPathErrors[cont.LogPath] = true
				log.WithField("err", errSize).Error("failed to get log file size")
			}
			continue
		}
		logs = append(logs, l)
		totalSize += l.Size
		delete(logPathErrors, cont.LogPath)
	}

	b, err := json.Marshal(struct {
		Logs      []*types.LogFileInfo
		TotalSize int64
	}{
		Logs:      logs,
		TotalSize: totalSize,
	})
	if err != nil {
		return fmt.Errorf("failed to encode as JSON: %w", err)
	}

	store.Set("dockerLogSize", b)
	return nil
}

func logFileSize(ci *dockerTypes.ContainerJSON, volumes []types.HostVolume) (*types.LogFileInfo, error) {
	var err error
	for _, vol := range volumes {
		p := path.Join(vol.Path, ci.LogPath)

		fi, statErr := os.Stat(p)
		if statErr != nil {
			err = statErr
			continue
		}

		return &types.LogFileInfo{
			ContainerID:   ci.ID,
			ContainerName: ci.Name,
			Path:          ci.LogPath,
			Size:          fi.Size(),
			LastCheck:     time.Now().UnixMilli(),
		}, nil
	}
	return nil, err // return last os.Stat error
}
