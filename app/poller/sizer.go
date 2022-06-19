package poller

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/amerkurev/doku/app/docker"
	"github.com/amerkurev/doku/app/types"
	"github.com/amerkurev/doku/app/util"
	dockerTypes "github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
)

func pathInfo(m dockerTypes.MountPoint, volumes []types.HostVolume) error {
	var err error
	for _, vol := range volumes {
		var p string
		if vol.Path == "/" {
			p = strings.TrimPrefix(m.Source, "/host_mnt")
		} else {
			p = strings.Replace(m.Source, "/host_mnt", vol.Path, 1)
		}

		fi, statErr := os.Stat(p)
		if statErr != nil {
			err = statErr
			continue
		}

		r := &types.HostPathInfo{
			Path:      m.Source,
			Size:      fi.Size(),
			ReadOnly:  !m.RW,
			LastCheck: time.Now().UnixMilli(),
		}
		if fi.IsDir() {
			size, files, e := util.DirSize(p)
			if e != nil {
				return e
			}
			r.Size = size
			r.IsDir = true
			r.Files = files
			r.LastCheck = time.Now().UnixMilli()
		}
		f := log.Fields{"path": r.Path, "size": r.Size, "ro": r.ReadOnly}
		log.WithFields(f).Debug("bind mounts")
		return nil
	}
	return err // return last os.Stat error
}

func mountsBindSize(ctx context.Context, d *docker.Client, volumes []types.HostVolume) {
	go func() {
		for {
			if containers, err := d.ContainerJSONList(ctx); err != nil {
				log.WithField("err", err).Error("failed to get the list of containers")
			} else {
				var seen []string

				for _, cont := range containers {
					if cont.Config != nil && contains("DOKU_IN_DOCKER=1", cont.Config.Env) {
						continue // skip myself
					}

					for _, m := range cont.Mounts {
						if m.Type == "bind" && !contains(m.Source, seen) {
							// prevent repeated getting a size
							seen = append(seen, m.Source)

							// get size of (bind) mounted directory
							if err := pathInfo(m, volumes); err != nil {
								f := log.Fields{"err": err, "mount": m}
								log.WithFields(f).Error("failed to get the mounted directory size")
							}

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
