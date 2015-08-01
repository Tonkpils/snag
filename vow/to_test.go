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

	e := "|\x1b[1;33mIn Progress\x1b[0m| snag: echo hello\r|\x1b[0;32mPassed\x1b[0m     |\nhello\n|\x1b[1;33mIn Progress\x1b[0m| snag: echo world\r|\x1b[0;32mPassed\x1b[0m     |\nworld\n"
	assert.Equal(t, e, testBuf.String())
	assert.True(t, result)
}

func TestExecCmdNotFound(t *testing.T) {
	var testBuf bytes.Buffer

	vow := To("echo", "hello")
	vow.Then("asdfasdf", "asdas")
	vow.Then("Shoud", "never", "happen")
	result := vow.Exec(&testBuf)

	e := "|\x1b[1;33mIn Progress\x1b[0m| snag: echo hello\r|\x1b[0;32mPassed\x1b[0m     |\nhello\n|\x1b[1;33mIn Progress\x1b[0m| snag: asdfasdf asdas\r|\x1b[0;31mFailed\x1b[0m     |\nexec: \"asdfasdf\": executable file not found in $PATH\n"
	assert.Equal(t, e, testBuf.String())
	assert.False(t, result)
}

func TestExecCmdFailed(t *testing.T) {
	var testBuf bytes.Buffer

	vow := To("echo", "hello")
	vow.Then("./test.sh")
	vow.Then("Shoud", "never", "happen")
	result := vow.Exec(&testBuf)

	e := "|\x1b[1;33mIn Progress\x1b[0m| snag: echo hello\r|\x1b[0;32mPassed\x1b[0m     |\nhello\n|\x1b[1;33mIn Progress\x1b[0m| snag: ./test.sh\r|\x1b[0;31mFailed\x1b[0m     |\n"
	assert.Equal(t, e, testBuf.String())
	assert.False(t, result)
}
