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

func TestNewBuilder_CmdWithQuotes(t *testing.T) {
	c := config{
		Build: []string{
			`echo "hello world" foo`,
			`echo "ga ga oh la la`,
			`echo "ga" "foo"`,
			`echo -c 'foo "bar"'`,
		},
	}

	b, err := NewBuilder(c)
	assert.NoError(t, err)

	assert.Equal(t, "echo", b.buildCmds[0][0])
	assert.Equal(t, `"hello world"`, b.buildCmds[0][1])
	assert.Equal(t, "foo", b.buildCmds[0][2])

	assert.Equal(t, "echo", b.buildCmds[1][0])
	assert.Equal(t, `"ga ga oh la la`, b.buildCmds[1][1])

	assert.Equal(t, "echo", b.buildCmds[2][0])
	assert.Equal(t, `"ga"`, b.buildCmds[2][1])
	assert.Equal(t, `"foo"`, b.buildCmds[2][2])

	assert.Equal(t, "echo", b.buildCmds[3][0])
	assert.Equal(t, `-c`, b.buildCmds[3][1])
	assert.Equal(t, `'foo "bar"'`, b.buildCmds[3][2])
}

func TestClose(t *testing.T) {
	b, err := NewBuilder(config{})
	require.NoError(t, err)

	err = b.Close()
	assert.NoError(t, err)

	_, ok := <-b.done
	assert.False(t, ok, "channel 'done' was not closed")
}
