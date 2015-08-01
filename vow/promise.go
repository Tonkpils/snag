package vow

import (
	"bytes"
	"io"
	"os/exec"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
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

func (p *promise) Run(w io.Writer) (result cmdResult, err error) {
	buf := new(bufCloser)
	p.cmd.Stdout = buf
	p.cmd.Stderr = buf

	// TODO: make the printing prettier
	w.Write([]byte("snag: " + strings.Join(p.cmd.Args, " ") + " - In Progress"))
	if err := p.cmd.Start(); err != nil {
		p.writeIfAlive(w, []byte("\b\b\b\b\b\b\b\b\b\b\bFailed       \n"))
		p.writeIfAlive(w, []byte(err.Error()+"\n"))
		return result, err
	}

	if err := p.cmd.Wait(); err != nil {
		result.failed = true
		p.writeIfAlive(w, []byte("\b\b\b\b\b\b\b\b\b\b\bFailed       \n"))
	} else {
		p.writeIfAlive(w, []byte("\b\b\b\b\b\b\b\b\b\b\bPassed       \n"))
	}

	p.writeIfAlive(w, buf.Bytes())

	result.ps = p.cmd.ProcessState
	result.command = p.cmd.Path
	result.args = p.cmd.Args[1:]
	return
}

func (p *promise) writeIfAlive(w io.Writer, b []byte) {
	if atomic.LoadInt32(p.killed) == 0 {
		w.Write([]byte(b))
	}
}

func (p *promise) kill() {
	atomic.StoreInt32(p.killed, 1)
	if p.cmd.Process != nil {
		p.cmd.Process.Signal(syscall.SIGTERM)

		for ; p.cmd.ProcessState != nil && !p.cmd.ProcessState.Exited(); <-time.After(100 * time.Millisecond) {
		}
	}
}
