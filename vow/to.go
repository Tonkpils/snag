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
	cancled *int32

	cmds []*promise
}

// To returns a new Vow that is configured to execute command given.
func To(name string, args ...string) *Vow {
	return &Vow{
		cmds:    []*promise{newPromise(name, args...)},
		cancled: new(int32),
	}
}

// Then adds the given command to the list of commands the Vow will execute
func (vow *Vow) Then(name string, args ...string) *Vow {
	vow.cmds = append(vow.cmds, newPromise(name, args...))
	return vow
}

func (vow *Vow) Close() {
	atomic.StoreInt32(vow.cancled, 1)
	for i := 0; i < len(vow.cmds); i++ {
		vow.cmds[i].stop()
	}
}

func (vow *Vow) isCancled() bool {
	return atomic.LoadInt32(vow.cancled) == 1
}

// Exec runs all of the commands a Vow has with all output redirected
// to the given writer and returns a Result
func (vow *Vow) Exec(w io.Writer) *Result {
	r := new(Result)
	var runCount int
	for !vow.isCancled() && runCount < len(vow.cmds) {
		result, err := vow.cmds[runCount].Run(w)
		if err != nil {
			// log.Fatal(err)
			result.failed = true
		}
		r.results = append(r.results, result)

		// manually increment the counter so that
		// in the case the command.failed is true
		// the loop below that adds an empty cmdResult
		// to the result doesn't add an extra result
		runCount++
		if result.failed {
			r.Failed = true
			break
		}
	}

	r.executed = runCount

	// add the remaining commands
	for ; runCount < len(vow.cmds); runCount++ {
		p := vow.cmds[runCount]
		command := cmdResult{
			command: p.cmd.Path,
			args:    p.cmd.Args,
		}
		r.results = append(r.results, command)
	}

	return r
}
