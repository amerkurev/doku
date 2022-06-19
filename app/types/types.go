package types

// HostVolume contains input information for a volume and the result for utilization percentage.
type HostVolume struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// LogFile

// HostPathInfo contains information about a filesystem path (file or directory).
type HostPathInfo struct {
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	IsDir     bool   `json:"isDir"`
	Files     int64  `json:"files"`
	ReadOnly  bool   `json:"readOnly"`
	LastCheck int64  `json:"lastCheck"`
}
