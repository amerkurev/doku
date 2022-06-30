package poller

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/amerkurev/doku/app/docker"
	"github.com/amerkurev/doku/app/store"
	"github.com/amerkurev/doku/app/types"
	"github.com/amerkurev/doku/app/util"
)

func bindMountsSize(ctx context.Context, d *docker.Client, volumes []types.HostVolume) {
	go func() {
		for {
			if containers, err := d.ContainerJSONList(ctx); err != nil {
				log.WithField("err", err).Error("failed to get the list of containers")
			} else {
				var seen []string
				bindMounts := make([]*types.BindMountInfo, 0)

				// gather all bind mounts
				for _, cont := range containers {
					if cont.Config != nil && contains("DOKU_IN_DOCKER=1", cont.Config.Env) {
						continue // skip myself (doku)
					}

					for _, m := range cont.Mounts {
						if m.Type == "bind" && !contains(m.Source, seen) {
							seen = append(seen, m.Source)
							bindMounts = append(bindMounts, &types.BindMountInfo{
								Path:      m.Source,
								ReadOnly:  !m.RW,
								LastCheck: time.Now().UnixMilli(),
							})
						}
					}
				}

				// for early access by API
				storeBindMounts(bindMounts, 0)

				var totalSize int64
				for _, m := range bindMounts {
					err := bindMountInfo(m, volumes) // get size of mounted directory (bind type)
					if err != nil {
						m.Err = err.Error()
						log.WithField("err", err).Error("failed to get mounted file or directory")
					} else {
						totalSize += m.Size
					}
					storeBindMounts(bindMounts, totalSize)

					// let the processor cool down
					if interruptionPoint(ctx, time.Second) {
						return
					}
				}
			}

			// return from function or pause the current goroutine for at least an hour
			if interruptionPoint(ctx, time.Hour) {
				return
			}
		}
	}()
}

func storeBindMounts(bindMounts []*types.BindMountInfo, totalSize int64) {
	b, err := json.Marshal(struct {
		BindMounts []*types.BindMountInfo
		TotalSize  int64
	}{
		BindMounts: bindMounts,
		TotalSize:  totalSize,
	})
	if err != nil {
		log.WithField("err", err).Error("failed to encode as JSON")
	}
	store.Set("dockerBindMounts", b)

}

func bindMountInfo(m *types.BindMountInfo, volumes []types.HostVolume) error {
	var err error
	for _, vol := range volumes {
		m.Prepared = true

		p := strings.TrimPrefix(m.Path, "/host_mnt")
		if !strings.HasPrefix(p, vol.Path) {
			p = path.Join(vol.Path, p)
		}

		fi, statErr := os.Stat(p)
		if statErr != nil {
			err = statErr
			continue
		}

		m.Size = fi.Size()
		m.LastCheck = time.Now().UnixMilli()

		if fi.IsDir() {
			size, files, e := util.DirSize(p)
			if e != nil {
				return e
			}
			m.Size = size
			m.IsDir = true
			m.Files = files
			m.LastCheck = time.Now().UnixMilli()
		}
		return nil
	}
	return err // return last os.Stat error
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
