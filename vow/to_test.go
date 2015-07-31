package vow

import (
	"bytes"
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

	e := []byte("snag: echo hello - In Progress\b\b\b\b\b\b\b\b\b\b\bPassed       \nhello\nsnag: echo world - In Progress\b\b\b\b\b\b\b\b\b\b\bPassed       \nworld\n")
	assert.Equal(t, e, testBuf.Bytes())
	assert.False(t, result.Failed)
	assert.Equal(t, result.executed, 2)
}

func TestExecCmdNotFound(t *testing.T) {
	var testBuf bytes.Buffer

	vow := To("echo", "hello")
	vow.Then("asdfasdf", "asdas")
	vow.Then("Shoud", "never", "happen")
	result := vow.Exec(&testBuf)

	e := []byte("snag: echo hello - In Progress\b\b\b\b\b\b\b\b\b\b\bPassed       \nhello\nsnag: asdfasdf asdas - In Progress\b\b\b\b\b\b\b\b\b\b\bFailed       \nexec: \"asdfasdf\": executable file not found in $PATH\n")
	assert.Equal(t, e, testBuf.Bytes())
	assert.True(t, result.Failed)
	assert.Equal(t, result.executed, 2)
	assert.Len(t, result.results, 3)
}

func TestExecCmdFailed(t *testing.T) {
	var testBuf bytes.Buffer

	vow := To("echo", "hello")
	vow.Then("./test.sh")
	vow.Then("Shoud", "never", "happen")
	result := vow.Exec(&testBuf)

	e := []byte("snag: echo hello - In Progress\b\b\b\b\b\b\b\b\b\b\bPassed       \nhello\nsnag: ./test.sh - In Progress\b\b\b\b\b\b\b\b\b\b\bFailed       \n")
	assert.Equal(t, e, testBuf.Bytes())
	assert.True(t, result.Failed)
	assert.Equal(t, result.executed, 2)
	assert.Len(t, result.results, 3)
}
