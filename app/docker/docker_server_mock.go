package docker

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ServerMock is a mock implementation of DockerServer.
type ServerMock interface {
	Start(t *testing.T)
	Shutdown(t *testing.T)
}

type serverMock struct {
	address string
	server  *http.Server
	done    chan struct{}
}

// NewMockServer creates the mock implementation of DockerServer.
func NewMockServer(addr, version, logPath, mountSource string) ServerMock {
	r := chi.NewRouter()
	prefix := "/" + version

	r.Get("/_ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK")) // nolint:gosec
	})
	r.Get(prefix+"/version", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(versionResponse)) // nolint:gosec
	})
	r.Get(prefix+"/system/df", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(diskUsageResponse)) // nolint:gosec
	})
	r.Get(prefix+"/containers/json", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(containerListResponse)) // nolint:gosec
	})
	r.Get(prefix+"/containers/"+containerID+"/json", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(containerInspectResponse(logPath, mountSource))) // nolint:gosec
	})
	r.Get(prefix+"/events", func(w http.ResponseWriter, r *http.Request) {
		for _, event := range eventsResponse {
			w.Write([]byte(event)) // nolint:gosec
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			time.Sleep(150 * time.Millisecond) // 150!
		}
	})

	s := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	return &serverMock{
		address: addr,
		server:  s,
		done:    make(chan struct{}, 1),
	}
}

// Start starts the mock implementation of DockerServer.
func (s *serverMock) Start(t *testing.T) {
	go func() {
		err := s.server.ListenAndServe()
		assert.ErrorIs(t, http.ErrServerClosed, err)
		s.done <- struct{}{}
	}()
}

// Shutdown shut down the mock implementation of DockerServer.
func (s *serverMock) Shutdown(t *testing.T) {
	err := s.server.Shutdown(context.Background())
	require.NoError(t, err)
	<-s.done
}

const versionResponse = `
{
  "Platform": {
    "Name": "Docker Engine - Community"
  },
  "Components": [
    {
      "Name": "Engine",
      "Version": "20.10.12",
      "Details": {
        "ApiVersion": "1.41",
        "Arch": "arm64",
        "BuildTime": "2021-12-13T11:43:07.000000000+00:00",
        "Experimental": "false",
        "GitCommit": "459d0df",
        "GoVersion": "go1.16.12",
        "KernelVersion": "5.10.76-linuxkit",
        "MinAPIVersion": "1.12",
        "Os": "linux"
      }
    },
    {
      "Name": "containerd",
      "Version": "1.4.12",
      "Details": {
        "GitCommit": "7b11cfaabd73bb80907dd23182b9347b4245eb5d"
      }
    },
    {
      "Name": "runc",
      "Version": "1.0.2",
      "Details": {
        "GitCommit": "v1.0.2-0-g52b36a2"
      }
    },
    {
      "Name": "docker-init",
      "Version": "0.19.0",
      "Details": {
        "GitCommit": "de40ad0"
      }
    }
  ],
  "Version": "20.10.12",
  "ApiVersion": "1.41",
  "MinAPIVersion": "1.12",
  "GitCommit": "459d0df",
  "GoVersion": "go1.16.12",
  "Os": "linux",
  "Arch": "arm64",
  "KernelVersion": "5.10.76-linuxkit",
  "BuildTime": "2021-12-13T11:43:07.000000000+00:00"
}`

const diskUsageResponse = `
{
  "LayersSize": 3790091975,
  "Images": [
    {
      "Containers": 0,
      "Created": 1655679047,
      "Id": "sha256:d295860a7c7407e010a41fe8e8fac27e88bffb37de32f42962925db7d1f0e5fc",
      "Labels": null,
      "ParentId": "",
      "RepoDigests": null,
      "RepoTags": [
        "amerkurev/doku:master"
      ],
      "SharedSize": 0,
      "Size": 9151574,
      "VirtualSize": 9151574
    }
  ],
  "Containers": [
    {
      "Id": "ff71fad88dcf97b382ff8ad56a0a1ed27ad2107c68126508460b60f2a53b4f12",
      "Names": [
        "/charming_chaplygin"
      ],
      "Image": "alpine:edge",
      "ImageID": "sha256:de97f69592930737615e9db075100faee03aa53e05d87ecfe1ceda16d1587ec6",
      "Command": "sh",
      "Created": 1655815545,
      "Ports": [],
      "SizeRootFs": 5349283,
      "Labels": {},
      "State": "running",
      "Status": "Up 3 minutes",
      "HostConfig": {
        "NetworkMode": "default"
      },
      "NetworkSettings": {
        "Networks": {
          "bridge": {
            "IPAMConfig": null,
            "Links": null,
            "Aliases": null,
            "NetworkID": "b01b0d3c7b9cfd320cb333167f7e7714affd9744265432d551c5c16f99ceeb56",
            "EndpointID": "f6e995f6d351ae987dacd4ba0f655d20bc09bc9f9127036708b1556acb6162cf",
            "Gateway": "172.17.0.1",
            "IPAddress": "172.17.0.2",
            "IPPrefixLen": 16,
            "IPv6Gateway": "",
            "GlobalIPv6Address": "",
            "GlobalIPv6PrefixLen": 0,
            "MacAddress": "02:42:ac:11:00:02",
            "DriverOpts": null
          }
        }
      },
      "Mounts": [
        {
          "Type": "bind",
          "Source": "/Users/username/data",
          "Destination": "/usr/src/data",
          "Mode": "",
          "RW": true,
          "Propagation": "rprivate"
        },
		{
          "Type": "bind",
          "Source": "/var/run/docker.sock",
          "Destination": "/var/run/docker.sock",
          "Mode": "",
          "RW": true,
          "Propagation": "rprivate"
        }
      ]
    }
  ],
  "Volumes": [
    {
      "CreatedAt": "2022-06-02T18:43:36Z",
      "Driver": "local",
      "Labels": null,
      "Mountpoint": "/var/lib/docker/volumes/34708689acbdaf656f5052ec2e155eaaaf7a8db81dfe30ddc871c6056696ebba/_data",
      "Name": "34708689acbdaf656f5052ec2e155eaaaf7a8db81dfe30ddc871c6056696ebba",
      "Options": null,
      "Scope": "local",
      "UsageData": {
        "RefCount": 0,
        "Size": 899710
      }
    }
  ],
  "BuildCache": [
    {
      "ID": "mfj9wygbwrmvrhuvw5lzxsnti",
      "Parent": "",
      "Type": "regular",
      "Description": "pulled from docker.io/library/alpine:3.14.1@sha256:eb3e4e175ba6d212ba1d6e04fc0782916c08e1c9d7b45892e9796141b1d379ae",
      "InUse": false,
      "Shared": true,
      "Size": 0,
      "CreatedAt": "2021-12-08T08:37:35.257551632Z",
      "LastUsedAt": "2022-01-31T10:48:45.397526833Z",
      "UsageCount": 2
    }
  ],
  "BuilderSize": 21809292023
}`

const containerListResponse = `
[
  {
    "Id": "ff71fad88dcf97b382ff8ad56a0a1ed27ad2107c68126508460b60f2a53b4f12",
    "Names": [
      "/charming_chaplygin"
    ],
    "Image": "alpine:edge",
    "ImageID": "sha256:de97f69592930737615e9db075100faee03aa53e05d87ecfe1ceda16d1587ec6",
    "Command": "sh",
    "Created": 1655815545,
    "Ports": [],
    "Labels": {},
    "State": "running",
    "Status": "Up 15 minutes",
    "HostConfig": {
      "NetworkMode": "default"
    },
    "NetworkSettings": {
      "Networks": {
        "bridge": {
          "IPAMConfig": null,
          "Links": null,
          "Aliases": null,
          "NetworkID": "b01b0d3c7b9cfd320cb333167f7e7714affd9744265432d551c5c16f99ceeb56",
          "EndpointID": "f6e995f6d351ae987dacd4ba0f655d20bc09bc9f9127036708b1556acb6162cf",
          "Gateway": "172.17.0.1",
          "IPAddress": "172.17.0.2",
          "IPPrefixLen": 16,
          "IPv6Gateway": "",
          "GlobalIPv6Address": "",
          "GlobalIPv6PrefixLen": 0,
          "MacAddress": "02:42:ac:11:00:02",
          "DriverOpts": null
        }
      }
    },
    "Mounts": [
      {
        "Type": "bind",
        "Source": "/Users/username/data",
        "Destination": "/usr/src/data",
        "Mode": "",
        "RW": true,
        "Propagation": "rprivate"
      },
      {
	    "Type": "bind",
	    "Source": "/var/run/docker.sock",
	    "Destination": "/var/run/docker.sock",
	    "Mode": "",
	    "RW": true,
	    "Propagation": "rprivate"
	  }
    ]
  }
]`

const containerID = "ff71fad88dcf97b382ff8ad56a0a1ed27ad2107c68126508460b60f2a53b4f12"
const containerInspectTemplate = `
{
  "Id": "ff71fad88dcf97b382ff8ad56a0a1ed27ad2107c68126508460b60f2a53b4f12",
  "Created": "2022-06-21T12:45:45.700912708Z",
  "Path": "sh",
  "Args": [],
  "State": {
    "Status": "running",
    "Running": true,
    "Paused": false,
    "Restarting": false,
    "OOMKilled": false,
    "Dead": false,
    "Pid": 2075,
    "ExitCode": 0,
    "Error": "",
    "StartedAt": "2022-06-21T12:45:46.034762208Z",
    "FinishedAt": "0001-01-01T00:00:00Z"
  },
  "Image": "sha256:de97f69592930737615e9db075100faee03aa53e05d87ecfe1ceda16d1587ec6",
  "ResolvConfPath": "/var/lib/docker/containers/ff71fad88dcf97b382ff8ad56a0a1ed27ad2107c68126508460b60f2a53b4f12/resolv.conf",
  "HostnamePath": "/var/lib/docker/containers/ff71fad88dcf97b382ff8ad56a0a1ed27ad2107c68126508460b60f2a53b4f12/hostname",
  "HostsPath": "/var/lib/docker/containers/ff71fad88dcf97b382ff8ad56a0a1ed27ad2107c68126508460b60f2a53b4f12/hosts",
  "LogPath": "%s",
  "Name": "/charming_chaplygin",
  "RestartCount": 0,
  "Driver": "overlay2",
  "Platform": "linux",
  "MountLabel": "",
  "ProcessLabel": "",
  "AppArmorProfile": "",
  "ExecIDs": null,
  "HostConfig": {
    "Binds": [
      "/Users/username/data:/usr/src/data"
    ],
    "ContainerIDFile": "",
    "LogConfig": {
      "Type": "json-file",
      "Config": {}
    },
    "NetworkMode": "default",
    "PortBindings": {},
    "RestartPolicy": {
      "Name": "no",
      "MaximumRetryCount": 0
    },
    "AutoRemove": true,
    "VolumeDriver": "",
    "VolumesFrom": null,
    "CapAdd": null,
    "CapDrop": null,
    "CgroupnsMode": "private",
    "Dns": [],
    "DnsOptions": [],
    "DnsSearch": [],
    "ExtraHosts": null,
    "GroupAdd": null,
    "IpcMode": "private",
    "Cgroup": "",
    "Links": null,
    "OomScoreAdj": 0,
    "PidMode": "",
    "Privileged": false,
    "PublishAllPorts": false,
    "ReadonlyRootfs": false,
    "SecurityOpt": null,
    "UTSMode": "",
    "UsernsMode": "",
    "ShmSize": 67108864,
    "Runtime": "runc",
    "ConsoleSize": [
      0,
      0
    ],
    "Isolation": "",
    "CpuShares": 0,
    "Memory": 0,
    "NanoCpus": 0,
    "CgroupParent": "",
    "BlkioWeight": 0,
    "BlkioWeightDevice": [],
    "BlkioDeviceReadBps": null,
    "BlkioDeviceWriteBps": null,
    "BlkioDeviceReadIOps": null,
    "BlkioDeviceWriteIOps": null,
    "CpuPeriod": 0,
    "CpuQuota": 0,
    "CpuRealtimePeriod": 0,
    "CpuRealtimeRuntime": 0,
    "CpusetCpus": "",
    "CpusetMems": "",
    "Devices": [],
    "DeviceCgroupRules": null,
    "DeviceRequests": null,
    "KernelMemory": 0,
    "KernelMemoryTCP": 0,
    "MemoryReservation": 0,
    "MemorySwap": 0,
    "MemorySwappiness": null,
    "OomKillDisable": null,
    "PidsLimit": null,
    "Ulimits": null,
    "CpuCount": 0,
    "CpuPercent": 0,
    "IOMaximumIOps": 0,
    "IOMaximumBandwidth": 0,
    "MaskedPaths": [
      "/proc/asound",
      "/proc/acpi",
      "/proc/kcore",
      "/proc/keys",
      "/proc/latency_stats",
      "/proc/timer_list",
      "/proc/timer_stats",
      "/proc/sched_debug",
      "/proc/scsi",
      "/sys/firmware"
    ],
    "ReadonlyPaths": [
      "/proc/bus",
      "/proc/fs",
      "/proc/irq",
      "/proc/sys",
      "/proc/sysrq-trigger"
    ]
  },
  "GraphDriver": {
    "Data": {
      "LowerDir": "/var/lib/docker/overlay2/cbc484b4831b8360dda89042d5ab380eeb3ce725aad65e184675006f2c42de5d-init/diff:/var/lib/docker/overlay2/29015c4f7a22932245e72e7af57469ee24b6cac0754e2495e89ad45a331c2dcc/diff",
      "MergedDir": "/var/lib/docker/overlay2/cbc484b4831b8360dda89042d5ab380eeb3ce725aad65e184675006f2c42de5d/merged",
      "UpperDir": "/var/lib/docker/overlay2/cbc484b4831b8360dda89042d5ab380eeb3ce725aad65e184675006f2c42de5d/diff",
      "WorkDir": "/var/lib/docker/overlay2/cbc484b4831b8360dda89042d5ab380eeb3ce725aad65e184675006f2c42de5d/work"
    },
    "Name": "overlay2"
  },
  "Mounts": [
    {
      "Type": "bind",
      "Source": "%s",
      "Destination": "/usr/src/data",
      "Mode": "",
      "RW": true,
      "Propagation": "rprivate"
    },
	{
	  "Type": "bind",
	  "Source": "/var/run/docker.sock",
	  "Destination": "/var/run/docker.sock",
	  "Mode": "",
	  "RW": true,
	  "Propagation": "rprivate"
	}
  ],
  "Config": {
    "Hostname": "ff71fad88dcf",
    "Domainname": "",
    "User": "",
    "AttachStdin": true,
    "AttachStdout": true,
    "AttachStderr": true,
    "Tty": true,
    "OpenStdin": true,
    "StdinOnce": true,
    "Env": [
      "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
    ],
    "Cmd": [
      "sh"
    ],
    "Image": "alpine:edge",
    "Volumes": null,
    "WorkingDir": "",
    "Entrypoint": null,
    "OnBuild": null,
    "Labels": {}
  },
  "NetworkSettings": {
    "Bridge": "",
    "SandboxID": "175400b9366b09750df6454bd30697fa0f9a2dbe0a6c7c3be48174cffe76ce99",
    "HairpinMode": false,
    "LinkLocalIPv6Address": "",
    "LinkLocalIPv6PrefixLen": 0,
    "Ports": {},
    "SandboxKey": "/var/run/docker/netns/175400b9366b",
    "SecondaryIPAddresses": null,
    "SecondaryIPv6Addresses": null,
    "EndpointID": "f6e995f6d351ae987dacd4ba0f655d20bc09bc9f9127036708b1556acb6162cf",
    "Gateway": "172.17.0.1",
    "GlobalIPv6Address": "",
    "GlobalIPv6PrefixLen": 0,
    "IPAddress": "172.17.0.2",
    "IPPrefixLen": 16,
    "IPv6Gateway": "",
    "MacAddress": "02:42:ac:11:00:02",
    "Networks": {
      "bridge": {
        "IPAMConfig": null,
        "Links": null,
        "Aliases": null,
        "NetworkID": "b01b0d3c7b9cfd320cb333167f7e7714affd9744265432d551c5c16f99ceeb56",
        "EndpointID": "f6e995f6d351ae987dacd4ba0f655d20bc09bc9f9127036708b1556acb6162cf",
        "Gateway": "172.17.0.1",
        "IPAddress": "172.17.0.2",
        "IPPrefixLen": 16,
        "IPv6Gateway": "",
        "GlobalIPv6Address": "",
        "GlobalIPv6PrefixLen": 0,
        "MacAddress": "02:42:ac:11:00:02",
        "DriverOpts": null
      }
    }
  }
}`

func containerInspectResponse(logPath, mountSource string) string {
	return fmt.Sprintf(containerInspectTemplate, logPath, mountSource)
}

var eventsResponse = [...]string{
	`{"status":"die","id":"ff71fad88dcf97b382ff8ad56a0a1ed27ad2107c68126508460b60f2a53b4f12","from":"alpine:edge","Type":"container","Action":"die","Actor":{"ID":"ff71fad88dcf97b382ff8ad56a0a1ed27ad2107c68126508460b60f2a53b4f12","Attributes":{"exitCode":"0","image":"alpine:edge","name":"charming_chaplygin"}},"scope":"local","time":1655818985,"timeNano":1655818985175023758}`,
	`{"Type":"network","Action":"disconnect","Actor":{"ID":"b01b0d3c7b9cfd320cb333167f7e7714affd9744265432d551c5c16f99ceeb56","Attributes":{"container":"ff71fad88dcf97b382ff8ad56a0a1ed27ad2107c68126508460b60f2a53b4f12","name":"bridge","type":"bridge"}},"scope":"local","time":1655818985,"timeNano":1655818985215693591}`,
	`{"status":"destroy","id":"ff71fad88dcf97b382ff8ad56a0a1ed27ad2107c68126508460b60f2a53b4f12","from":"alpine:edge","Type":"container","Action":"destroy","Actor":{"ID":"ff71fad88dcf97b382ff8ad56a0a1ed27ad2107c68126508460b60f2a53b4f12","Attributes":{"image":"alpine:edge","name":"charming_chaplygin"}},"scope":"local","time":1655818985,"timeNano":1655818985230071050}`,
	`{"status":"create","id":"6a7a7c7547049d41acdb009324395a5067bef169a7421a1a189e63477e4bde20","from":"alpine:edge","Type":"container","Action":"create","Actor":{"ID":"6a7a7c7547049d41acdb009324395a5067bef169a7421a1a189e63477e4bde20","Attributes":{"desktop.docker.io/binds/0/Source":"/Users/username/data","desktop.docker.io/binds/0/SourceKind":"hostFile","desktop.docker.io/binds/0/Target":"/usr/src/data","image":"alpine:edge","name":"pensive_pasteur"}},"scope":"local","time":1655818996,"timeNano":1655818996536216680}`,
	`{"status":"attach","id":"6a7a7c7547049d41acdb009324395a5067bef169a7421a1a189e63477e4bde20","from":"alpine:edge","Type":"container","Action":"attach","Actor":{"ID":"6a7a7c7547049d41acdb009324395a5067bef169a7421a1a189e63477e4bde20","Attributes":{"desktop.docker.io/binds/0/Source":"/Users/username/data","desktop.docker.io/binds/0/SourceKind":"hostFile","desktop.docker.io/binds/0/Target":"/usr/src/data","image":"alpine:edge","name":"pensive_pasteur"}},"scope":"local","time":1655818996,"timeNano":1655818996544098889}`,
	`{"Type":"network","Action":"connect","Actor":{"ID":"b01b0d3c7b9cfd320cb333167f7e7714affd9744265432d551c5c16f99ceeb56","Attributes":{"container":"6a7a7c7547049d41acdb009324395a5067bef169a7421a1a189e63477e4bde20","name":"bridge","type":"bridge"}},"scope":"local","time":1655818996,"timeNano":1655818996606499097}`,
	`{"status":"start","id":"6a7a7c7547049d41acdb009324395a5067bef169a7421a1a189e63477e4bde20","from":"alpine:edge","Type":"container","Action":"start","Actor":{"ID":"6a7a7c7547049d41acdb009324395a5067bef169a7421a1a189e63477e4bde20","Attributes":{"desktop.docker.io/binds/0/Source":"/Users/username/data","desktop.docker.io/binds/0/SourceKind":"hostFile","desktop.docker.io/binds/0/Target":"/usr/src/data","image":"alpine:edge","name":"pensive_pasteur"}},"scope":"local","time":1655818996,"timeNano":1655818996778273347}`,
	`{"status":"resize","id":"6a7a7c7547049d41acdb009324395a5067bef169a7421a1a189e63477e4bde20","from":"alpine:edge","Type":"container","Action":"resize","Actor":{"ID":"6a7a7c7547049d41acdb009324395a5067bef169a7421a1a189e63477e4bde20","Attributes":{"desktop.docker.io/binds/0/Source":"/Users/username/data","desktop.docker.io/binds/0/SourceKind":"hostFile","desktop.docker.io/binds/0/Target":"/usr/src/data","height":"49","image":"alpine:edge","name":"pensive_pasteur","width":"180"}},"scope":"local","time":1655818996,"timeNano":1655818996782988389}`,
}
