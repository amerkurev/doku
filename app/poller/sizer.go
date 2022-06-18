package poller

import (
	"context"
	"github.com/amerkurev/doku/app/docker"
	"github.com/amerkurev/doku/app/store"
	"os"
	"path"
	"time"

	"github.com/amerkurev/doku/app/types"
	dockerTypes "github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
)

func dirSizeCalculator(ctx context.Context, d *docker.Client, volumes []types.HostVolume) {
	go func() {
		for {
			// return from function or pause the current goroutine for at least 1 ms
			if interruptionPoint(ctx) {
				return
			}

			containers, err := d.ContainerList(ctx, dockerTypes.ContainerListOptions{All: true})
			if err != nil {
				log.WithField("err", err).Error("failed to get the list of containers in the docker host")
				continue
			}

			for _, c := range containers {
				ci, err := d.ContainerInspect(ctx, c.ID)
				if err != nil {
					break
				}

				for _, vol := range volumes {
					p := path.Join(vol.Path, ci.LogPath)
					if fi, err := os.Stat(p); err == nil {
						r := &types.HostPathInfo{
							Path:      ci.LogPath,
							Size:      fi.Size(),
							LastCheck: time.Now().UnixMilli(),
						}
						store.Set(ci.LogPath, r)
						break
					}
				}

				if interruptionPoint(ctx) {
					return
				}
			}

			// docker bind mounts info
			// for _, path := range docker.BindMounts(r.DiskUsage.Containers) {
			//	size, files, err := util.DirSize(path)
			//	if err != nil {
			//		fmt.Printf("err: %+v", err)
			//		continue
			//	}
			//	fmt.Printf("%s, %d bytes, %d files\n", path, size, files)
			// }
		}
	}()
}

func interruptionPoint(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	case <-time.After(time.Second):
		return false
	}
}
