package builder

import (
	"io"
	"sync/atomic"
)

// vow represents a batch of commands being prepared to run
type vow struct {
	canceled *int32

	cmds    []*promise
	Verbose bool
}

// VowTo returns a new vow that is configured to execute command given.
func VowTo(name string, args ...string) *vow {
	return &vow{
		cmds:     []*promise{newPromise(name, args...)},
		canceled: new(int32),
	}
}

// Then adds the given command to the list of commands the Vow will execute
func (v *vow) Then(name string, args ...string) *vow {
	v.cmds = append(v.cmds, newPromise(name, args...))
	return v
}

func (v *vow) ThenAsync(name string, args ...string) *vow {
	v.cmds = append(v.cmds, newAsyncPromise(name, args...))
	return v
}

// Stop terminates the active command and stops the execution of any future commands
func (v *vow) Stop() {
	atomic.StoreInt32(v.canceled, 1)
	for i := 0; i < len(v.cmds); i++ {
		v.cmds[i].kill()
	}
}

func (v *vow) isCanceled() bool {
	return atomic.LoadInt32(v.canceled) == 1
}

// Exec runs all of the commands a Vow has with all output redirected
// to the given writer and returns a Result
func (v *vow) Exec(w io.Writer) bool {
	for i := 0; i < len(v.cmds); i++ {
		if v.isCanceled() {
			return false
		}

		if err := v.cmds[i].Run(w, v.Verbose); err != nil {
			return false
		}
	}
	return true
}
