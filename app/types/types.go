package types

import (
	"github.com/shirou/gopsutil/v3/disk"
)

// HostVolume contains input information for a volume and the result for utilization percentage.
type HostVolume struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	disk.UsageStat `json:"usageStat"`
}

// CopyFrom fill HostVolume struct from gopsutil.disk.UsageStat.
func (v *HostVolume) CopyFrom(du *disk.UsageStat) {
	v.UsageStat.Path = du.Path
	v.Fstype = du.Fstype
	v.Total = du.Total
	v.Free = du.Free
	v.Used = du.Used
	v.UsedPercent = du.UsedPercent
	v.InodesTotal = du.InodesTotal
	v.InodesUsed = du.InodesUsed
	v.InodesFree = du.InodesFree
	v.InodesUsedPercent = du.InodesUsedPercent
}
