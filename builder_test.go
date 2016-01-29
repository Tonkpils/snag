package main

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuilder(t *testing.T) {
	testEnv := "foobar"
	os.Setenv("TEST_ENV", testEnv)
	c := config{
		Build:        []string{"echo Hello World", "echo $$TEST_ENV"},
		Run:          []string{"echo async here"},
		IgnoredItems: []string{"foo", "bar"},
		Verbose:      true,
	}
	b, err := NewBuilder(c)
	assert.NoError(t, err)
	assert.NotNil(t, b)

	require.Len(t, b.buildCmds, 2)
	assert.Equal(t, c.Build[0], strings.Join(b.buildCmds[0], " "))
	assert.Equal(t, testEnv, b.buildCmds[1][1])

	require.Len(t, b.runCmds, 1)
	assert.Equal(t, c.Run[0], strings.Join(b.runCmds[0], " "))

	assert.Equal(t, c.Verbose, b.verbose)
	assert.Equal(t, c.IgnoredItems, b.ignoredItems)
}

func TestClose(t *testing.T) {
	b, err := NewBuilder(config{})
	require.NoError(t, err)

	err = b.Close()
	assert.NoError(t, err)

	_, ok := <-b.done
	assert.False(t, ok, "channel 'done' was not closed")
}
