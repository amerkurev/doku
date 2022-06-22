package http

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/amerkurev/doku/app/store"
)

func Test_Server_Run(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	err := store.Initialize()
	require.NoError(t, err)

	port := 1000 + rand.Intn(1000)
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	allowed := []string{
		"user:$2y$10$ViWb72wAg4nEKdHh.9p2yeEKLSN0EBkbZ0Mf0bqNHZmItsQt6K8he",
	}

	httpServer := &Server{
		Address: addr,
		Timeouts: Timeouts{
			Read:        5 * time.Second,
			Write:       60 * time.Second,
			Idle:        60 * time.Second,
			Shutdown:    5 * time.Second,
			LongPolling: 30 * time.Second, // it must be less than writeTimeout!
		},
		BasicAuthEnabled: true,
		BasicAuthAllowed: allowed,
		StaticFolder:     "../../frontend/static",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		err := httpServer.Run(ctx)
		assert.NoError(t, err)
		done <- struct{}{}
	}()

	tbl := []struct {
		endpoint string
	}{
		{"/"},
		{"/favicon.ico"},
		{"/version"},
		{"/size-calc-progress"},
		{"/docker/version"},
		{"/docker/disk-usage"},
		{"/docker/log-info"},
		{"/docker/mounts-bind"},
		// {"/docker/_/docker/disk-usage"},
	}

	time.Sleep(10 * time.Millisecond)

	client := http.Client{}

	for i, tt := range tbl {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			req, err := http.NewRequest("GET", "http://127.0.0.1:"+strconv.Itoa(port)+tt.endpoint, http.NoBody)
			require.NoError(t, err)
			req.SetBasicAuth("user", "1111")
			resp, err := client.Do(req)
			require.NoError(t, err)
			err = resp.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}

	cancel()
	<-done
}

func Test_Server_RunFailed(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	err := store.Initialize()
	require.NoError(t, err)

	port := 1_000_000
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	httpServer := &Server{
		Address: addr,
		Timeouts: Timeouts{
			Read:        5 * time.Second,
			Write:       60 * time.Second,
			Idle:        60 * time.Second,
			Shutdown:    5 * time.Second,
			LongPolling: 30 * time.Second, // it must be less than writeTimeout!
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		err := httpServer.Run(ctx)
		assert.Error(t, err)
		assert.EqualError(t, errors.New("http server failed: listen tcp: address 1000000: invalid port"), err.Error())
		done <- struct{}{}
	}()

	<-done
}
