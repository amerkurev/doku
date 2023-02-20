package store

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type payload struct {
	Name string
}

func Test_Store(t *testing.T) {
	key := "key"
	wrongKey := "wrong-key"

	d := payload{"payload"}
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
}
