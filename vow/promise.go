package vow

import (
	"io"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type promise struct {
	cmd    *exec.Cmd
	closed chan struct{}
}

func newPromise(name string, args ...string) *promise {
	return &promise{
		cmd:    exec.Command(name, args...),
		closed: make(chan struct{}),
	}
}

func (p *promise) Run(w io.Writer) (result cmdResult, err error) {
	defer close(p.closed)

	cw := newCdmWriter(w)
	defer cw.Close()

	p.cmd.Stdout = cw
	p.cmd.Stderr = cw

	w.Write([]byte(strings.Join(p.cmd.Args, " ") + "\n"))
	if err := p.cmd.Start(); err != nil {
		w.Write([]byte(err.Error() + "\n"))
		return result, err
	}

	go func() {
		<-p.closed
		cw.Close()
	}()

	if err := p.cmd.Wait(); err != nil {
		result.failed = true
	}

	result.ps = p.cmd.ProcessState
	result.command = p.cmd.Path
	result.args = p.cmd.Args[1:]
	return
}

func (p *promise) stop() {
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
