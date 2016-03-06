package builder

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
	"time"
)

var errKilled = errors.New("promise has already been killed")

var (
	statusFailed     = "\r|" + red("Failed") + "     |\n"
	statusPassed     = "\r|" + green("Passed") + "     |\n"
	statusInProgress = "|" + yellow("In Progress") + "|"
)

type syncBuffer struct {
	sync.RWMutex

	buf *bytes.Buffer
}

func newSyncBuffer() *syncBuffer {
	return &syncBuffer{buf: bytes.NewBuffer([]byte{})}
}

func (sb *syncBuffer) Write(p []byte) (int, error) {
	sb.Lock()
	n, err := sb.buf.Write(p)
	sb.Unlock()
	return n, err
}

func (sb *syncBuffer) Read(p []byte) (int, error) {
	sb.RLock()
	n, err := sb.buf.Read(p)
	sb.RUnlock()
	return n, err
}

func (sb *syncBuffer) Next(n int) []byte {
	sb.RLock()
	b := sb.buf.Next(n)
	sb.RUnlock()
	return b
}

func (sb *syncBuffer) Bytes() []byte {
	sb.RLock()
	b := sb.buf.Bytes()
	sb.RUnlock()
	return b
}

type promise struct {
	cmdMtx sync.Mutex
	cmd    *exec.Cmd
	async  bool
	killed *int32
}

func newPromise(name string, args ...string) *promise {
	return &promise{
		cmd:    exec.Command(name, args...),
		killed: new(int32),
	}
}

func newAsyncPromise(name string, args ...string) *promise {
	p := newPromise(name, args...)
	p.async = true
	return p
}

func (p *promise) Run(w io.Writer, verbose bool) (err error) {
	if p.isKilled() {
		return errKilled
	}

	buf := newSyncBuffer()
	p.cmd.Stdout = buf
	p.cmd.Stderr = buf

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

	// if the process is async we don't need to do anything else
	if p.async {
		fmt.Println(" -- process id: ", p.cmd.Process.Pid)
		go p.fowardOutput(p.cmd.Process.Pid, w, buf)
		go p.wait(w, verbose, buf)
		return nil
	}

	return p.wait(w, verbose, buf)
}

func (p *promise) wait(w io.Writer, verbose bool, buf *syncBuffer) error {
	p.cmdMtx.Lock()
	err := p.cmd.Wait()
	p.cmdMtx.Unlock()

	status := statusPassed
	if err != nil {
		status = statusFailed
	}

	if p.async {
		status = status[1 : len(status)-1]
		status = fmt.Sprintf("%s %s\n", status, strings.Join(p.cmd.Args, " "))
	}

	p.writeIfAlive(w, []byte(status))

	if verbose || err != nil {
		p.writeIfAlive(w, buf.Bytes())
	}

	return err
}

func (p *promise) fowardOutput(pid int, w io.Writer, buf *syncBuffer) {
	prefix := []byte(yellow(fmt.Sprintf("pid %d : ", pid)))
	for t := time.Tick(time.Second); !p.isKilled(); <-t {
		b := buf.Next(1024)
		if len(b) == 0 {
			continue
		}

		p.writeIfAlive(w, append(prefix, b...))
	}
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
