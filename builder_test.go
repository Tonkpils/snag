package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuilder(t *testing.T) {
	b, err := NewBuilder(config{})
	assert.NoError(t, err)
	assert.NotNil(t, b)
}

func TestNewBuilder_EnvScript(t *testing.T) {
	testEnv := "foobar"
	os.Setenv("TEST_ENV", testEnv)
	b, err := NewBuilder(config{
		Build: []string{"echo $$TEST_ENV"},
	})
	assert.NoError(t, err)
	assert.NotNil(t, b)

	require.Len(t, b.cmds, 1)
	assert.Equal(t, testEnv, b.cmds[0][1])
}

func TestClose(t *testing.T) {
	b, err := NewBuilder(config{})
	require.NoError(t, err)

	err = b.Close()
	assert.NoError(t, err)

	_, ok := <-b.done
	assert.False(t, ok, "channel 'done' was not closed")
}
