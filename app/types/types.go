package types

// AppVersion contains version of the Doku app.
type AppVersion struct {
	Version string `json:"version"`
}

// HostVolume contains input information for a volume and the result for utilization percentage.
type HostVolume struct {
	Name string
	Path string
}

// LogFileInfo contains information about a container log file.
type LogFileInfo struct {
	ContainerID   string
	ContainerName string
	Path          string
	Size          int64
	LastCheck     int64
}

// HostPathInfo contains information about a filesystem path (file or directory).
type HostPathInfo struct {
	Path      string
	Size      int64
	IsDir     bool
	Files     int64
	ReadOnly  bool
	LastCheck int64
}
