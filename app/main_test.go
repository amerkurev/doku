package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDummy(t *testing.T) {
	assert.Equal(t, dummy(1, 2), 3)
}
