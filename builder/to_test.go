package builder

import (
	"testing"
	"time"

	"github.com/Tonkpils/snag/exchange"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	echoScript = "../fixtures/echo.sh"
	failScript = "../fixtures/fail.sh"
)

type TestExchange struct{}

func (ts *TestExchange) Send(event string, data interface{}) {}

func (ts *TestExchange) Listen(event string, fn func(ev exchange.Event)) {
}

func TestTo(t *testing.T) {
	cmd := "foo"
	args := []string{"Hello", "Worlf!"}
	vow := VowTo(&TestExchange{}, cmd, args...)
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
	v := VowTo(&TestExchange{}, echoScript)
	for i := 0; i < 50; i++ {
		v = v.Then(echoScript)
	}

	result := make(chan bool)
	defer close(result)

	started := make(chan struct{})
	go func() {
		close(started)
		result <- v.Exec()
	}()
	<-started

	v.Stop()
	assert.True(t, v.isCanceled())

	r := <-result
	assert.False(t, r)
}

func TestStopAsync(t *testing.T) {
	v := VowTo(&TestExchange{}, echoScript)
	v.ThenAsync(echoScript)

	require.True(t, v.Exec())
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
	v := VowTo(&TestExchange{}, echoScript)
	v.Then(echoScript)

	result := v.Exec()

	assert.True(t, result)
}

func TestExecCmdNotFound(t *testing.T) {
	v := VowTo(&TestExchange{}, echoScript)
	v.Then("asdfasdf", "asdas")
	v.Then("Shoud", "never", "happen")

	result := v.Exec()

	assert.False(t, result)
}

func TestExecCmdFailed(t *testing.T) {
	v := VowTo(&TestExchange{}, echoScript)
	v.Then(failScript)
	v.Then("Shoud", "never", "happen")

	result := v.Exec()

	assert.False(t, result)
}

func TestVowVerbose(t *testing.T) {
	v := VowTo(&TestExchange{}, echoScript)
	v.Verbose = true

	result := v.Exec()

	assert.True(t, result)
}
