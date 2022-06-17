package util

import "github.com/shirou/gopsutil/v3/disk"

// Volume contains input information for a volume and the result for utilization percentage.
type Volume struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	disk.UsageStat `json:"usageStat"`
}
