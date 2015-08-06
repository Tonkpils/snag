package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuilder(t *testing.T) {
	b, err := NewBuilder(config{})
	assert.NoError(t, err)
	assert.NotNil(t, b)
}

func TestClose(t *testing.T) {
	b, err := NewBuilder(config{})
	require.NoError(t, err)

	b.Close()

	_, ok := <-b.done
	assert.False(t, ok, "channel 'done' was not closed")
}
