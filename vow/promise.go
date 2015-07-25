package vow

import (
	"io"
	"log"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

var pollTime = 50 * time.Millisecond

type promise struct {
	cmd *exec.Cmd
}

func newPromise(name string, args ...string) *promise {
	return &promise{
		cmd: exec.Command(name, args...),
	}
}

func (p *promise) Run(w io.Writer) (result cmdResult, err error) {
	cw := newCdmWriter(w)
	defer cw.Close()

	p.cmd.Stdout = cw
	p.cmd.Stderr = cw

	w.Write([]byte(strings.Join(p.cmd.Args, " ") + "\n"))
	if err := p.cmd.Start(); err != nil {
		w.Write([]byte(err.Error() + "\n"))
		return result, err
	}

	go p.monitorClose(cw)

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

		for ; p.cmd.ProcessState != nil && !p.cmd.ProcessState.Exited(); <-time.After(pollTime) {
		}
	}
}

func (p *promise) monitorClose(w io.WriteCloser) {
	for ; p.cmd.ProcessState == nil; time.After(pollTime) {
	}

	for ; ; <-time.After(pollTime) {
		status := p.cmd.ProcessState.Sys().(syscall.WaitStatus)
		switch {
		case status.Signaled():
			w.Close()
			return
		case status.Exited(), status.Stopped():
			return
		default:
			log.Printf("signal %#v\n", status)
		}
	}
}
