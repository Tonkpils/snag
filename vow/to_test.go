package vow

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	echoScript = "./echo.sh"
	failScript = "./fail.sh"
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

	vow := To(echoScript)
	vow.Then(echoScript)
	result := vow.Exec(&testBuf)

	e := fmt.Sprintf(
		"%s %s%s%s %s%s",
		statusInProgress,
		echoScript,
		statusPassed,
		statusInProgress,
		echoScript,
		statusPassed,
	)
	assert.Equal(t, e, testBuf.String())
	assert.True(t, result)
}

func TestExecCmdNotFound(t *testing.T) {
	var testBuf bytes.Buffer

	vow := To(echoScript)
	vow.Then("asdfasdf", "asdas")
	vow.Then("Shoud", "never", "happen")
	result := vow.Exec(&testBuf)

	e := fmt.Sprintf(
		"%s %s%s%s asdfasdf asdas%sexec: \"asdfasdf\": executable file not found in ",
		statusInProgress,
		echoScript,
		statusPassed,
		statusInProgress,
		statusFailed,
	)

	assert.True(t, strings.HasPrefix(testBuf.String(), e))
	assert.False(t, result)
}

func TestExecCmdFailed(t *testing.T) {
	var testBuf bytes.Buffer

	vow := To(echoScript)
	vow.Then(failScript)
	vow.Then("Shoud", "never", "happen")
	result := vow.Exec(&testBuf)

	e := fmt.Sprintf(
		"%s %s%s%s %s%s",
		statusInProgress,
		echoScript,
		statusPassed,
		statusInProgress,
		failScript,
		statusFailed,
	)

	assert.Equal(t, e, testBuf.String())
	assert.False(t, result)
}

func TestVowVerbose(t *testing.T) {
	var testBuf bytes.Buffer

	vow := To(echoScript)
	vow.Verbose = true
	result := vow.Exec(&testBuf)
	e := fmt.Sprintf(
		"%s %s%shello\r\n",
		statusInProgress,
		echoScript,
		statusPassed,
	)

	assert.Equal(t, e, testBuf.String())
	assert.True(t, result)
}
