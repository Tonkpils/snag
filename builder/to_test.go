package builder

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
	vow := VowTo(cmd, args...)
	require.NotNil(t, vow)

	assert.Len(t, vow.cmds, 1)
	assert.Equal(t, cmd, vow.cmds[0].cmd.Path)
	assert.Equal(t, args, vow.cmds[0].cmd.Args[1:])
}

func TestThen(t *testing.T) {
	var v vow
	totalCmds := 10
	for i := 0; i < totalCmds; i++ {
		v.Then("foo", "bar", "baz")
	}
	v.Then("foo").Then("another")

	assert.Len(t, v.cmds, totalCmds+2)
}

func TestThenAsync(t *testing.T) {
	var v vow
	v.ThenAsync("foo", "bar", "baz")

	require.Len(t, v.cmds, 1)
	assert.True(t, v.cmds[0].async)
}

func TestStop(t *testing.T) {
	v := VowTo(echoScript)
	for i := 0; i < 50; i++ {
		v = v.Then(echoScript)
	}

	result := make(chan bool)
	defer close(result)

	started := make(chan struct{})
	go func() {
		close(started)
		result <- v.Exec(ioutil.Discard)
	}()
	<-started

	v.Stop()
	assert.True(t, v.isCanceled())

	r := <-result
	assert.False(t, r)
}

func TestStopAsync(t *testing.T) {
	v := VowTo(echoScript)
	v.ThenAsync(echoScript)

	require.True(t, v.Exec(ioutil.Discard))
	<-time.After(10 * time.Millisecond)

	v.Stop()
	for _, p := range v.cmds {
		p.cmdMtx.Lock()
		p.cmd.Wait()
		assert.True(t, p.cmd.ProcessState.Exited())
		p.cmdMtx.Unlock()
	}
}

func TestExec(t *testing.T) {
	var testBuf bytes.Buffer

	v := VowTo(echoScript)
	v.Then(echoScript)
	result := v.Exec(&testBuf)

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

	v := VowTo(echoScript)
	v.Then("asdfasdf", "asdas")
	v.Then("Shoud", "never", "happen")
	result := v.Exec(&testBuf)

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

	v := VowTo(echoScript)
	v.Then(failScript)
	v.Then("Shoud", "never", "happen")
	result := v.Exec(&testBuf)

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

	v := VowTo(echoScript)
	v.Verbose = true
	result := v.Exec(&testBuf)
	e := fmt.Sprintf(
		"%s %s%shello\r\n",
		statusInProgress,
		echoScript,
		statusPassed,
	)

	assert.Equal(t, e, testBuf.String())
	assert.True(t, result)
}
