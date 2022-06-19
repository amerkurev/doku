package poller

import (
	"context"
	"encoding/json"
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

var mounts = make(map[string]*types.HostPathInfo)

type progress struct {
	Start    time.Time `json:"start"`
	Duration int64     `json:"duration"`
	Done     bool      `json:"done"`
}

func mountsBindSize(ctx context.Context, d *docker.Client, volumes []types.HostVolume) {
	go func() {
		for {
			op := progress{Start: time.Now()}

			if containers, err := d.ContainerJSONList(ctx); err != nil {
				log.WithField("err", err).Error("failed to get the list of containers")
			} else {
				var seen []string

				for _, cont := range containers {
					if cont.Config != nil && contains("DOKU_IN_DOCKER=1", cont.Config.Env) {
						continue // skip myself (doku)
					}

					for _, m := range cont.Mounts {
						if m.Type == "bind" && !contains(m.Source, seen) {
							// prevent repeated getting a size
							seen = append(seen, m.Source)

							// get size of mounted directory (bind type)
							pi, err := pathInfo(m, volumes)
							if err != nil {
								log.WithField("err", err).Error("failed to get mounted file or directory")
							} else {
								mounts[m.Source] = pi
								b, err := json.Marshal(mounts)
								if err != nil {
									log.WithField("err", err).Error("failed to encode as JSON")
								} else {
									store.Set("dockerMountsBind", b) // for early access by API
								}
							}

							// let the processor cool down
							if interruptionPoint(ctx, time.Second) {
								return
							}
						}
					}
				}
			}

			op.Duration = time.Since(op.Start).Milliseconds()
			op.Done = true
			b, err := json.Marshal(op)
			if err != nil {
				log.WithField("err", err).Error("failed to encode as JSON")
			}
			store.Set("sizeCalcProgress", b)
			log.WithField("took", op.Duration).Debug("size calc progress")

			// return from function or pause the current goroutine for at least 5 minutes
			if interruptionPoint(ctx, 5*time.Minute) {
				return
			}
		}
	}()
}

func pathInfo(m dockerTypes.MountPoint, volumes []types.HostVolume) (*types.HostPathInfo, error) {
	var err error
	for _, vol := range volumes {
		p := strings.TrimPrefix(m.Source, "/host_mnt")
		if !strings.HasPrefix(p, vol.Path) {
			p = path.Join(vol.Path, p)
		}

		fi, statErr := os.Stat(p)
		if statErr != nil {
			err = statErr
			continue
		}

		res := &types.HostPathInfo{
			Path:      m.Source,
			Size:      fi.Size(),
			ReadOnly:  !m.RW,
			LastCheck: time.Now().UnixMilli(),
		}
		if fi.IsDir() {
			size, files, e := util.DirSize(p)
			if e != nil {
				return nil, e
			}
			res.Size = size
			res.IsDir = true
			res.Files = files
			res.LastCheck = time.Now().UnixMilli()
		}
		return res, nil
	}
	return nil, err // return last os.Stat error
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
