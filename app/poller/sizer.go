package poller

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"sort"
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
						continue // skip myself (doku container)
					}

					for _, m := range cont.Mounts {
						if m.Source == "/var/run/docker.sock" {
							continue // skip Docker unix socket (for all)
						}

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
					err := fillData(m, volumes) // get size of mounted directory (bind type)
					if err != nil {
						m.Err = err.Error()
						log.WithField("err", err).Error("failed to get mounted file or directory")
					} else {
						totalSize = bindMountsTotalSize(bindMounts)
					}
					storeBindMounts(bindMounts, totalSize)

					// let the cpu cool down
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

func fillData(m *types.BindMountInfo, volumes []types.HostVolume) error {
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

func bindMountsTotalSize(bindMounts []*types.BindMountInfo) int64 {
	var mounts []*types.BindMountInfo

	for _, m := range bindMounts {
		if m.Size > 0 {
			mounts = append(mounts, m)
		}
	}

	length := len(mounts)

	if length == 0 {
		return 0
	}

	if length == 1 {
		return mounts[0].Size
	}

	sort.Slice(mounts, func(i, j int) bool {
		return mounts[i].Path < mounts[j].Path
	})

	paths := []string{mounts[0].Path}
	totalSize := mounts[0].Size

	for i := 1; i < length; i++ {
		if !strings.HasPrefix(mounts[i].Path, paths[len(paths)-1]) {
			paths = append(paths, mounts[i].Path)
			totalSize += mounts[i].Size
		}
	}

	return totalSize
}
