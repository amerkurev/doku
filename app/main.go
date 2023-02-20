// Package main provides the entry point for the application.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"

	"github.com/amerkurev/doku/app/bindmount"
	"github.com/amerkurev/doku/app/docker"
	"github.com/amerkurev/doku/app/http"
	"github.com/amerkurev/doku/app/types"
)

var opts struct {
	Listen            string   `short:"l" long:"listen" env:"LISTEN" description:"listen on host:port (default: 0.0.0.0:9090 under docker, 127.0.0.1:9090 without)"`
	AuthBasicHtpasswd string   `long:"basic-htpasswd" env:"BASIC_HTPASSWD" description:"htpasswd file for basic auth"`
	Volumes           []string `short:"v" long:"volume" env:"VOLUMES" default:"root:/hostroot" env-delim:"," description:"volumes to report"`

	Docker struct {
		Host     string `long:"host" env:"HOST" default:"unix:///var/run/docker.sock" description:"url to the docker server"`
		CertPath string `long:"cert" env:"CERT_PATH" description:"path to the TLS certificates"`
		Verify   bool   `long:"verify" env:"TLS_VERIFY" description:"enable or disable TLS verification, off by default"`
		Version  string `long:"version" env:"API_VERSION" description:"version of the API to reach, leave empty for latest"`
	} `group:"docker" namespace:"docker" env-namespace:"DOCKER"`

	UI struct {
		Home   string `long:"home" env:"HOME" default:"web/static" description:"path to the location of the static folder"`
		Title  string `long:"title" env:"TITLE" default:"Docker disk usage" description:"title of the document"`
		Header string `long:"header" env:"HEADER" default:"Docker disk space usage" description:"header at the top of the dashboard"`
	} `group:"ui" namespace:"ui" env-namespace:"UI"`

	Log struct {
		StdOut bool   `long:"stdout" env:"STDOUT" description:"enable stdout logging"`
		Level  string `long:"level" env:"LEVEL" description:"logging level" choice:"debug" choice:"info" choice:"warning" choice:"error" default:"info"`
	} `group:"log" namespace:"log" env-namespace:"LOG"`
}

var revision = "unknown"

func main() {
	fmt.Printf("doku %s\n", revision)

	parser := flags.NewParser(&opts, flags.Default)
	parser.SubcommandsOptional = true

	if _, err := parser.Parse(); err != nil {
		os.Exit(2)
	}

	configureLogging(opts.Log.Level, opts.Log.StdOut)

	volumes, err := parseVolumes(opts.Volumes)
	if err != nil {
		log.WithField("err", err).Fatal("parse volume failed")
	}

	if err := run(volumes); err != nil {
		log.WithField("err", err).Fatal("doku failed")
	}
	log.Info("goodbye")
}

func run(volumes []types.HostVolume) error {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, types.CtxKeyRevision, revision)
	ctx = context.WithValue(ctx, types.CtxKeyVolumes, volumes)

	go func() {
		// catch signal and invoke graceful termination
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
		<-signalChannel
		log.Debug("interrupt signal")
		cancel()
	}()

	d, err := docker.NewClient(
		ctx,
		opts.Docker.Host,
		opts.Docker.CertPath,
		opts.Docker.Version,
		opts.Docker.Verify,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize docker client: %w", err)
	}
	log.Info(fmt.Sprintf("docker engine API (%s)", d.ClientVersion()))

	log.Info("calculating size of volumes...")
	bindmount.CalcSize(ctx, d, volumes)

	basicAuthAllowed, err := makeBasicAuth(opts.AuthBasicHtpasswd)
	if err != nil {
		return fmt.Errorf("failed to load basic auth: %w", err)
	}

	addr := listenAddress(opts.Listen)
	httpServer := &http.Server{
		Address: addr,
		Timeouts: http.Timeouts{
			Read:     5 * time.Second,
			Write:    60 * time.Second,
			Idle:     60 * time.Second,
			Shutdown: 5 * time.Second,
		},
		BasicAuthEnabled: len(basicAuthAllowed) > 0,
		BasicAuthAllowed: basicAuthAllowed,
		StaticFolder:     opts.UI.Home,
		UITitle:          opts.UI.Title,
		UIHeader:         opts.UI.Header,
	}

	log.Info(fmt.Sprintf("starting http server at %s", addr))
	return httpServer.Run(ctx, d)
}

func configureLogging(level string, stdout bool) {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	if stdout {
		log.SetOutput(os.Stdout)
	}

	switch level {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warning":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	}
}

// listenAddress sets default to 127.0.0.1:9090 and, if detected DOKU_IN_DOCKER env, to 0.0.0.0:9090
func listenAddress(addr string) string {

	// don't set default if any opts.Listen address defined by user
	if addr != "" {
		return addr
	}

	// http, set default to 9090 in docker, 80 without
	if v, ok := os.LookupEnv("DOKU_IN_DOCKER"); ok && (v == "1" || v == "true") {
		return "0.0.0.0:9090"
	}
	return "127.0.0.1:9090"
}

// makeBasicAuth returns a list of allowed basic auth users and password hashes.
// if no htpasswd file is specified, an empty list is returned.
func makeBasicAuth(htpasswdFile string) ([]string, error) {
	var basicAuthAllowed []string
	if htpasswdFile != "" {
		data, err := os.ReadFile(htpasswdFile) //nolint:gosec //read file with opts passed path
		if err != nil {
			return nil, fmt.Errorf("failed to read htpasswd file %s: %w", htpasswdFile, err)
		}
		basicAuthAllowed = strings.Split(string(data), "\n")
		for i, v := range basicAuthAllowed {
			basicAuthAllowed[i] = strings.TrimSpace(v)
			basicAuthAllowed[i] = strings.Replace(basicAuthAllowed[i], "\t", "", -1)
		}
	}
	return basicAuthAllowed, nil
}

// parseVolumes parses volumes from string list, each element in format "name:path"
func parseVolumes(volumes []string) ([]types.HostVolume, error) {
	res := make([]types.HostVolume, len(volumes))
	for i, v := range volumes {
		parts := strings.SplitN(v, ":", 2)
		if len(parts) != 2 {
			return nil, errors.New("invalid volume format, should be <name>:<path>")
		}
		res[i] = types.HostVolume{Name: parts[0], Path: parts[1]}
		log.WithField("path", parts[1]).Debug("volume to report")
	}
	return res, nil
}
