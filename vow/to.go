/*
Package vow provides a promise like api for executing
a batch of external commands
*/
package vow

import (
	"io"
	"sync/atomic"
)

// Vow represents a batch of commands being prepared to run
type Vow struct {
	canceled *int32

	cmds    []*promise
	Verbose bool
}

// To returns a new Vow that is configured to execute command given.
func To(name string, args ...string) *Vow {
	return &Vow{
		cmds:     []*promise{newPromise(name, args...)},
		canceled: new(int32),
	}
}

// Then adds the given command to the list of commands the Vow will execute
func (vow *Vow) Then(name string, args ...string) *Vow {
	vow.cmds = append(vow.cmds, newPromise(name, args...))
	return vow
}

// Stop terminates the active command and stops the execution of any future commands
func (vow *Vow) Stop() {
	atomic.StoreInt32(vow.canceled, 1)
	for i := 0; i < len(vow.cmds); i++ {
		vow.cmds[i].kill()
	}
}

func (vow *Vow) isCanceled() bool {
	return atomic.LoadInt32(vow.canceled) == 1
}

// Exec runs all of the commands a Vow has with all output redirected
// to the given writer and returns a Result
func (vow *Vow) Exec(w io.Writer) bool {
	for i := 0; i < len(vow.cmds); i++ {
		if vow.isCanceled() {
			return false
		}

		if err := vow.cmds[i].Run(w, vow.Verbose); err != nil {
			return false
		}
	}
	return true
}
