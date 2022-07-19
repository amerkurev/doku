package util

import "syscall"

// DiskUsage contains usage data.
type DiskUsage struct {
	Total     uint64  // total size of the file system
	Used      uint64  // total bytes used in file system
	Free      uint64  // total free bytes on file system
	Available uint64  // total available bytes on file system to an unprivileged user
	Percent   float64 // percentage of use on the file system
	Path      string
}

// NewDiskUsage returns the DiskUsage instance of path or nil in case of error.
func NewDiskUsage(path string) (*DiskUsage, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil, err
	}

	du := &DiskUsage{
		Total:     stat.Blocks * uint64(stat.Bsize),
		Free:      stat.Bfree * uint64(stat.Bsize),
		Available: stat.Bavail * uint64(stat.Bsize),
		Path:      path,
	}

	du.Used = du.Total - du.Free
	du.Percent = 100 * float64(du.Used) / float64(du.Total)
	return du, nil
}
