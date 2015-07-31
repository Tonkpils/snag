package vow

import (
	"bytes"
	"io"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type promise struct {
	cmd  *exec.Cmd
	stop chan struct{}
}

func newPromise(name string, args ...string) *promise {
	return &promise{
		cmd:  exec.Command(name, args...),
		stop: make(chan struct{}),
	}
}

func (p *promise) Run(w io.Writer) (result cmdResult, err error) {
	defer close(p.stop)

	var b bytes.Buffer

	p.cmd.Stdout = &b
	p.cmd.Stderr = &b

	// TODO: make the printing prettier
	w.Write([]byte("snag: " + strings.Join(p.cmd.Args, " ") + " - In Progress"))
	if err := p.cmd.Start(); err != nil {
		w.Write([]byte("\b\b\b\b\b\b\b\b\b\b\bFailed       \n"))
		w.Write([]byte(err.Error() + "\n"))
		return result, err
	}

	if err := p.cmd.Wait(); err != nil {
		result.failed = true
		w.Write([]byte("\b\b\b\b\b\b\b\b\b\b\bFailed       \n"))
	} else {

		w.Write([]byte("\b\b\b\b\b\b\b\b\b\b\bPassed       \n"))
	}

	w.Write(b.Bytes())

	result.ps = p.cmd.ProcessState
	result.command = p.cmd.Path
	result.args = p.cmd.Args[1:]
	return
}

func (p *promise) kill() {
	if p.cmd.Process != nil {
		w, ok := p.cmd.Stdout.(io.ReadCloser)
		if ok {
			w.Close()
		}

		w, ok = p.cmd.Stdin.(io.ReadCloser)
		if ok {
			w.Close()
		}

		p.cmd.Process.Signal(syscall.SIGTERM)

		for ; p.cmd.ProcessState != nil && !p.cmd.ProcessState.Exited(); <-time.After(100 * time.Millisecond) {
		}
	}
}
