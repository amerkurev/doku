package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amerkurev/doku/app/docker"
	"github.com/amerkurev/doku/app/http"
	"github.com/amerkurev/doku/app/poller"
	"github.com/amerkurev/doku/app/store"
	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
)

var opts struct {
	Listen            string `short:"l" long:"listen" env:"LISTEN" description:"listen on host:port (default: 0.0.0.0:9090 under docker, 127.0.0.1:80 without)"`
	AuthBasicHtpasswd string `long:"basic-htpasswd" env:"BASIC_HTPASSWD" description:"htpasswd file for basic auth"`

	Docker struct {
		Host     string `long:"host" env:"HOST" default:"unix:///var/run/docker.sock" description:"url to the docker server"`
		CertPath string `long:"cert" env:"CERT_PATH" description:"path to the TLS certificates"`
		Verify   bool   `long:"verify" env:"TLS_VERIFY" description:"enable or disable TLS verification, off by default"`
		Version  string `long:"version" env:"API_VERSION" description:"version of the API to reach, leave empty for latest"`
	} `group:"docker" namespace:"docker" env-namespace:"DOCKER"`

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

	configureLogging()

	if err := run(); err != nil {
		log.WithField("err", err).Fatal("doku failed")
	}
	log.Info("goodbye")
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		// catch signal and invoke graceful termination
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
		<-signalChannel
		log.Debug("interrupt signal")
		cancel()
	}()

	err := store.Initialize()
	if err != nil {
		return fmt.Errorf("failed to initialize the store: %w", err)
	}
	store.Get().Set("revision", revision)

	d, err := docker.NewClient(opts.Docker.Host, opts.Docker.CertPath, opts.Docker.Version, opts.Docker.Verify)
	if err != nil {
		return fmt.Errorf("failed to initialize docker client: %w", err)
	}

	log.Info("starting docker poller")
	poller.Run(ctx, d)

	addr := listenAddress(opts.Listen)
	log.Info(fmt.Sprintf("starting http server at %s", addr))
	return http.Run(ctx, addr)
}

func configureLogging() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	if opts.Log.StdOut {
		log.SetOutput(os.Stdout)
	}

	switch opts.Log.Level {
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

// listenAddress sets default to 127.0.0.1:80 and, if detected DOKU_IN_DOCKER env, to 0.0.0.0:9090
func listenAddress(addr string) string {

	// don't set default if any opts.Listen address defined by user
	if addr != "" {
		return addr
	}

	// http, set default to 9090 in docker, 80 without
	if v, ok := os.LookupEnv("DOKU_IN_DOCKER"); ok && (v == "1" || v == "true") {
		return "0.0.0.0:9090"
	}
	return "127.0.0.1:80"
}
