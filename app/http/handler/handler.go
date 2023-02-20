// Package handler provides HTTP handlers for the application.
package handler

import (
	"context"
	"encoding/json"
	"net/http"
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

func internalServerError(w http.ResponseWriter, err error, reason string) {
	log.WithField("err", err).Error(reason)
	w.WriteHeader(http.StatusInternalServerError)
}

// Version returns version of the application.
func Version(ctx context.Context) http.HandlerFunc {
	revision := ctx.Value(types.CtxKeyRevision).(string)

	return func(w http.ResponseWriter, _ *http.Request) {
		b, err := json.Marshal(types.AppVersion{Version: revision})
		if err != nil {
			internalServerError(w, err, "failed to encode as JSON: version")
			return
		}

		w.Write(b) // nolint:gosec
	}
}

// DockerVersion returns version of the docker daemon.
func DockerVersion(ctx context.Context, d *docker.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		res, err := d.ServerVersion(ctx)
		if err != nil {
			internalServerError(w, err, "failed to execute request: docker version")
			return
		}

		// hide sensitive (and not only) data
		res.Components = nil

		b, err := json.Marshal(res)
		if err != nil {
			internalServerError(w, err, "failed to encode as JSON: docker version")
			return
		}

		w.Write(b) // nolint:gosec
	}
}

// DockerContainerList returns list of containers.
func DockerContainerList(ctx context.Context, d *docker.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		res, err := d.ContainerJSONList(ctx)
		if err != nil {
			internalServerError(w, err, "failed to execute request: docker container list")
			return
		}

		var totalSize int64
		for _, c := range res {
			// hide sensitive (and not only) data
			c.HostConfig = nil
			c.GraphDriver.Data = nil
			c.GraphDriver.Name = ""
			c.NetworkSettings = nil
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
			internalServerError(w, err, "failed to encode as JSON: docker container list")
			return
		}

		w.Write(b) // nolint:gosec
	}
}

// DockerDiskUsage returns disk usage of the docker daemon.
func DockerDiskUsage(ctx context.Context, d *docker.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		res, err := d.DiskUsage(ctx)
		if err != nil {
			internalServerError(w, err, "failed to execute request: docker disk usage")
			return
		}

		// hide sensitive (and not only) data
		res.Containers = nil

		// prevent null value for JSON array fields
		if res.Images == nil {
			res.Images = make([]*dockerTypes.ImageSummary, 0)
		}
		if res.Volumes == nil {
			res.Volumes = make([]*dockerTypes.Volume, 0)
		}
		if res.BuildCache == nil {
			res.BuildCache = make([]*dockerTypes.BuildCache, 0)
		}

		b, err := json.Marshal(res)
		if err != nil {
			internalServerError(w, err, "failed to encode as JSON: docker disk usage")
			return
		}

		w.Write(b) // nolint:gosec
	}
}

// DockerLogSize returns size of the docker container logs.
func DockerLogSize(ctx context.Context, d *docker.Client) http.HandlerFunc {
	volumes := ctx.Value(types.CtxKeyVolumes).([]types.HostVolume)
	logPathErrors := make(map[string]bool)

	return func(w http.ResponseWriter, _ *http.Request) {
		containers, err := d.ContainerJSONList(ctx)
		if err != nil {
			internalServerError(w, err, "failed to execute request: docker log size")
			return
		}

		logs := make([]*types.LogFileInfo, 0)
		var totalSize int64

		for _, cont := range containers {
			if len(cont.LogPath) == 0 {
				continue
			}

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
			internalServerError(w, err, "failed to encode as JSON: docker log size")
			return
		}

		w.Write(b) // nolint:gosec
	}
}

// DockerBindMounts returns size of bind mounts.
func DockerBindMounts() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		v, ok := store.Get("dockerBindMounts") // cached value
		if !ok {
			v = []byte("{}")
		}
		w.Write(v.([]byte)) // nolint:gosec
	}
}

// DiskUsage returns disk usage of the host.
func DiskUsage(ctx context.Context) http.HandlerFunc {
	volumes := ctx.Value(types.CtxKeyVolumes).([]types.HostVolume)

	return func(w http.ResponseWriter, _ *http.Request) {
		var lastErr error
		res := &util.DiskUsage{}

		for _, vol := range volumes {
			du, err := util.NewDiskUsage(vol.Path)
			if err != nil {
				lastErr = err
				continue
			}
			if du.Total > res.Total {
				res = du // the largest volume is taken
			}
		}

		if lastErr != nil && res.Total == 0 {
			internalServerError(w, lastErr, "failed to get disk usage")
			return
		}

		b, err := json.Marshal(res)
		if err != nil {
			internalServerError(w, err, "failed to encode as JSON: disk usage")
			return
		}

		w.Write(b) // nolint:gosec
	}
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
