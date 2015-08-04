package vow

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTo(t *testing.T) {
	cmd := "foo"
	args := []string{"Hello", "Worlf!"}
	vow := To(cmd, args...)
	require.NotNil(t, vow)

	assert.Len(t, vow.cmds, 1)
	assert.Equal(t, cmd, vow.cmds[0].cmd.Path)
	assert.Equal(t, args, vow.cmds[0].cmd.Args[1:])
}

func TestThen(t *testing.T) {
	var vow Vow
	totalCmds := 10
	for i := 0; i < totalCmds; i++ {
		vow.Then("foo", "bar", "baz")
	}
	vow.Then("foo").Then("another")

	assert.Len(t, vow.cmds, totalCmds+2)
}

func TestExec(t *testing.T) {
	var testBuf bytes.Buffer

	vow := To("echo", "hello")
	vow.Then("echo", "world")
	result := vow.Exec(&testBuf)

	e := fmt.Sprintf(
		"%s snag: echo hello%shello\n%s snag: echo world%sworld\n",
		statusInProgress,
		statusPassed,
		statusInProgress,
		statusPassed,
	)
	assert.Equal(t, e, testBuf.String())
	assert.True(t, result)
}

func TestExecCmdNotFound(t *testing.T) {
	var testBuf bytes.Buffer

	vow := To("echo", "hello")
	vow.Then("asdfasdf", "asdas")
	vow.Then("Shoud", "never", "happen")
	result := vow.Exec(&testBuf)

	e := fmt.Sprintf(
		"%s snag: echo hello%shello\n%s snag: asdfasdf asdas%sexec: \"asdfasdf\": executable file not found in $PATH\n",
		statusInProgress,
		statusPassed,
		statusInProgress,
		statusFailed,
	)
	assert.Equal(t, e, testBuf.String())
	assert.False(t, result)
}

func TestExecCmdFailed(t *testing.T) {
	var testBuf bytes.Buffer

	vow := To("echo", "hello")
	vow.Then("./test.sh")
	vow.Then("Shoud", "never", "happen")
	result := vow.Exec(&testBuf)

	e := fmt.Sprintf(
		"%s snag: echo hello%shello\n%s snag: ./test.sh%s",
		statusInProgress,
		statusPassed,
		statusInProgress,
		statusFailed,
	)

	assert.Equal(t, e, testBuf.String())
	assert.False(t, result)
}
