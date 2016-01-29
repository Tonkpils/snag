package vow

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	echoScript = "../fixtures/echo.sh"
	failScript = "../fixtures/fail.sh"
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

func TestThenAsync(t *testing.T) {
	var vow Vow
	vow.ThenAsync("foo", "bar", "baz")

	require.Len(t, vow.cmds, 1)
	assert.True(t, vow.cmds[0].async)
}

func TestStop(t *testing.T) {
	vow := To(echoScript)
	for i := 0; i < 50; i++ {
		vow = vow.Then(echoScript)
	}

	result := make(chan bool)
	defer close(result)

	started := make(chan struct{})
	go func() {
		close(started)
		result <- vow.Exec(ioutil.Discard)
	}()
	<-started

	vow.Stop()
	assert.True(t, vow.isCanceled())

	r := <-result
	assert.False(t, r)
}

func TestStopAsync(t *testing.T) {
	vow := To(echoScript)
	vow.ThenAsync(echoScript)

	require.True(t, vow.Exec(ioutil.Discard))
	<-time.After(10 * time.Millisecond)

	vow.Stop()
	for _, cmd := range vow.cmds {
		assert.True(t, cmd.cmd.ProcessState.Exited())
	}
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
