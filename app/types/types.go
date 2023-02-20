// Package types contains types used in the Doku app.
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

// BindMountInfo contains information about a bind mount.
type BindMountInfo struct {
	Path      string
	Size      int64
	IsDir     bool
	Files     int64
	ReadOnly  bool
	LastCheck int64
	Prepared  bool
	Err       string
}

// CtxKey is the type for the context keys.
type CtxKey string

const (
	// CtxKeyRevision is the key for the revision in the context.
	CtxKeyRevision CtxKey = "revision"
	// CtxKeyVolumes is the key for the volumes in the context.
	CtxKeyVolumes CtxKey = "volumes"
)
