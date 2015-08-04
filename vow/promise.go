package vow

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	statusFailed     = "\r|" + red("Failed") + "     |\n"
	statusPassed     = "\r|" + green("Passed") + "     |\n"
	statusInProgress = "|" + yellow("In Progress") + "|"
)

type bufCloser struct {
	bytes.Buffer
}

func (bc *bufCloser) Close() error {
	bc.Reset()
	return nil
}

type promise struct {
	cmd    *exec.Cmd
	killed *int32
}

func newPromise(name string, args ...string) *promise {
	return &promise{
		cmd:    exec.Command(name, args...),
		killed: new(int32),
	}
}

func (p *promise) Run(w io.Writer) (err error) {
	buf := new(bufCloser)
	p.cmd.Stdout = buf
	p.cmd.Stderr = buf

	fmt.Fprintf(
		w,
		"%s snag: %s",
		statusInProgress,
		strings.Join(p.cmd.Args, " "),
	)
	if err := p.cmd.Start(); err != nil {
		p.writeIfAlive(w, []byte(statusFailed))
		p.writeIfAlive(w, []byte(err.Error()+"\n"))
		return err
	}

	err = p.cmd.Wait()
	if err != nil {
		p.writeIfAlive(w, []byte(statusFailed))
	} else {
		p.writeIfAlive(w, []byte(statusPassed))
	}

	p.writeIfAlive(w, buf.Bytes())
	return err
}

func (p *promise) writeIfAlive(w io.Writer, b []byte) {
	if atomic.LoadInt32(p.killed) == 0 {
		w.Write([]byte(b))
	}
}

func (p *promise) kill() {
	atomic.StoreInt32(p.killed, 1)
	if p.cmd.Process != nil {
		if p.cmd.ProcessState != nil && !p.cmd.ProcessState.Exited() {
			p.cmd.Process.Signal(syscall.SIGTERM)
		}

		for ; p.cmd.ProcessState != nil &&
			!p.cmd.ProcessState.Exited(); <-time.After(100 * time.Millisecond) {
		}
	}
}
