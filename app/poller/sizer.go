package poller

import (
	"context"
	"os"
	"path"
	"strings"
	"time"

	"github.com/amerkurev/doku/app/docker"
	"github.com/amerkurev/doku/app/store"
	"github.com/amerkurev/doku/app/types"
	"github.com/amerkurev/doku/app/util"
	dockerTypes "github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
)

func containerLogFile(ci *dockerTypes.ContainerJSON, volumes []types.HostVolume) {
	for _, vol := range volumes {
		p := path.Join(vol.Path, ci.LogPath)
		if fi, err := os.Stat(p); err == nil {
			store.Set(ci.LogPath, &types.HostPathInfo{
				Path:      ci.LogPath,
				Size:      fi.Size(),
				LastCheck: time.Now().UnixMilli(),
			})
			log.WithFields(log.Fields{"path": ci.LogPath, "size": fi.Size()}).Debug("container log file")
			return
		}
	}
}

func bindMountedDirectory(m dockerTypes.MountPoint, volumes []types.HostVolume) {
	for _, vol := range volumes {
		var p string
		if vol.Path == "/" {
			p = strings.TrimPrefix(m.Source, "/host_mnt")
		} else {
			p = strings.Replace(m.Source, "/host_mnt", vol.Path, 1)
		}
		if fi, err := os.Stat(p); err == nil {
			if fi.IsDir() {
				size, files, err := util.DirSize(p)
				if err == nil {
					store.Set(m.Source, &types.HostPathInfo{
						Path:      m.Source,
						Size:      size,
						IsDir:     true,
						Files:     files,
						OnlyRead:  !m.RW,
						LastCheck: time.Now().UnixMilli(),
					})
					f := log.Fields{"path": m.Source, "size": size, "rw": m.RW}
					log.WithFields(f).Debug("(bind) mounted dir")
					return
				}
			} else {
				store.Set(m.Source, &types.HostPathInfo{
					Path:      m.Source,
					Size:      fi.Size(),
					OnlyRead:  !m.RW,
					LastCheck: time.Now().UnixMilli(),
				})
				f := log.Fields{"path": m.Source, "size": fi.Size(), "rw": m.RW}
				log.WithFields(f).Debug("(bind) mounted file")
				return
			}
		}
	}
}

func dirSizeCalculator(ctx context.Context, d *docker.Client, volumes []types.HostVolume) {
	go func() {
		for {
			if containers, err := d.ContainerList(ctx, dockerTypes.ContainerListOptions{All: true}); err != nil {
				log.WithField("err", err).Error("failed to get the list of containers")
			} else {
				var seen []string

				for _, c := range containers {
					ci, err := d.ContainerInspect(ctx, c.ID)
					if err != nil {
						log.WithField("err", err).Error("failed to inspect the container")
						break
					}

					if contains("DOKU_IN_DOCKER=1", ci.Config.Env) {
						continue // skip myself
					}

					// get size of container log file
					containerLogFile(&ci, volumes)
					// let the processor cool down
					if interruptionPoint(ctx, time.Second) {
						return
					}

					for _, m := range c.Mounts {
						if m.Type == "bind" && !contains(m.Source, seen) {
							// prevent repeated getting a size
							seen = append(seen, m.Source)
							// get size of (bind) mounted directory
							bindMountedDirectory(m, volumes)
							// let the processor cool down
							if interruptionPoint(ctx, time.Second) {
								return
							}
						}
					}
				}
			}

			// return from function or pause the current goroutine for at least 5 minutes
			if interruptionPoint(ctx, 5*time.Minute) {
				return
			}
		}
	}()
}

func interruptionPoint(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return true
	case <-time.After(d):
		return false
	}
}

func contains[T comparable](val T, list []T) bool {
	for _, v := range list {
		if val == v {
			return true
		}
	}
	return false
}
