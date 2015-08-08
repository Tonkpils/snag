package vow

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
)

var errKilled = errors.New("promise has already been killed")

var (
	statusFailed     = "\r|" + red("Failed") + "     |\n"
	statusPassed     = "\r|" + green("Passed") + "     |\n"
	statusInProgress = "|" + yellow("In Progress") + "|"
)

type promise struct {
	cmdMtx sync.Mutex
	cmd    *exec.Cmd
	killed *int32
}

func newPromise(name string, args ...string) *promise {
	return &promise{
		cmd:    exec.Command(name, args...),
		killed: new(int32),
	}
}

func (p *promise) Run(w io.Writer, verbose bool) (err error) {
	if p.isKilled() {
		return errKilled
	}

	var buf bytes.Buffer
	p.cmd.Stdout = &buf
	p.cmd.Stderr = &buf

	fmt.Fprintf(
		w,
		"%s %s",
		statusInProgress,
		strings.Join(p.cmd.Args, " "),
	)

	p.cmdMtx.Lock()
	if err := p.cmd.Start(); err != nil {
		p.cmdMtx.Unlock()
		p.writeIfAlive(w, []byte(statusFailed))
		p.writeIfAlive(w, []byte(err.Error()+"\n"))
		return err
	}
	p.cmdMtx.Unlock()

	err = p.cmd.Wait()

	status := statusPassed
	if err != nil {
		status = statusFailed
	}
	p.writeIfAlive(w, []byte(status))

	if verbose || err != nil {
		p.writeIfAlive(w, buf.Bytes())
	}
	return err
}

func (p *promise) writeIfAlive(w io.Writer, b []byte) {
	if p.isKilled() {
		return
	}
	// ignoring error since there is not much we can do
	_, _ = w.Write(b)
}

func (p *promise) isKilled() bool {
	return atomic.LoadInt32(p.killed) == 1
}

func (p *promise) kill() {
	atomic.StoreInt32(p.killed, 1)
	p.cmdMtx.Lock()
	if p.cmd.Process != nil {
		// if we can't signal the process assume it has died
		_ = p.cmd.Process.Signal(syscall.SIGTERM)
	}
	p.cmdMtx.Unlock()
}
