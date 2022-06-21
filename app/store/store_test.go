package store

import (
	"context"
	"io/ioutil"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type payload struct {
	Name string
}

func recoverUninitializedStore(t *testing.T) {
	r := recover()
	assert.IsType(t, r, new(log.Entry))
	entry := r.(*log.Entry)
	assert.Equal(t, errStoreUninitialized, entry.Message)
}

func Test_Store_UninitializedGet(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	defer recoverUninitializedStore(t)
	Get("key")
}

func Test_Store_UninitializedSet(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	defer recoverUninitializedStore(t)
	Set("key", struct{}{})
}

func Test_Store_UninitializedWait(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	defer recoverUninitializedStore(t)
	Wait(context.Background(), time.Minute)
}

func Test_Store_UninitializedNotifyAll(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	defer recoverUninitializedStore(t)
	NotifyAll()
}

func Test_Store(t *testing.T) {
	err := Initialize()
	require.NoError(t, err)
	err = Initialize()
	require.NoError(t, err)

	key := "some-key"
	wrongKey := "non-existent-key"

	d := payload{"username"}
	Set(key, d)

	v, ok := Get(wrongKey)
	assert.False(t, ok)
	assert.Nil(t, v)

	v, ok = Get(key)
	assert.True(t, ok)
	assert.NotNil(t, v)

	data, ok := v.(payload)
	assert.True(t, ok)
	assert.Equal(t, data, d)

	go func() {
		time.Sleep(time.Millisecond * 500)
		NotifyAll()
	}()

	ctx := context.Background()
	start := time.Now()
	ch := Wait(ctx, time.Minute)
	assert.Equal(t, struct{}{}, <-ch)
	assert.Less(t, time.Since(start), time.Second)

	var counter int64
	var wg sync.WaitGroup
	goroutines := 5
	wg.Add(goroutines)

	for n := 0; n < goroutines; n++ {
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				Set(strconv.Itoa(i), int64(i))
				time.Sleep(time.Millisecond)
				num, ok := Get(strconv.Itoa(i))
				assert.True(t, ok)
				atomic.AddInt64(&counter, num.(int64))
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, int64(24750), counter)
}
