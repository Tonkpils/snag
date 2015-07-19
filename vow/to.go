/*
Package vow provides a promise like api for executing
a batch of external commands
*/
package vow

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Vow represents a batch of commands being prepared to run
type Vow struct {
	cmds   []*exec.Cmd
	errlog *log.Logger
}

// To returns a new Vow that is configured to execute command given.
func To(name string, args ...string) *Vow {
	cmd := exec.Command(name, args...)
	return &Vow{
		cmds:   []*exec.Cmd{cmd},
		errlog: log.New(os.Stderr, "To - ", log.Lshortfile),
	}
}

// Then adds the given command to the list of commands the Vow will execute
func (vow *Vow) Then(name string, args ...string) {
	vow.cmds = append(vow.cmds, exec.Command(name, args...))
}

// Exec runs all of the commands a Vow has with all output redirected
// to the given writer and returns a Result
func (vow *Vow) Exec(w io.Writer) *Result {
	r := new(Result)
	var runCount int
	for runCount < len(vow.cmds) {
		command := vow.runCmd(vow.cmds[runCount], w)
		r.results = append(r.results, command)

		// manually increment the counter so that
		// in the case the command.failed is true
		// the loop below that adds an empty cmdResult
		// to the result doesn't add an extra result
		runCount++
		if command.failed {
			r.Failed = true
			break
		}
	}

	r.executed = runCount

	// add the remaining commands
	for ; runCount < len(vow.cmds); runCount++ {
		cmd := vow.cmds[runCount]
		command := cmdResult{
			command: cmd.Path,
			args:    cmd.Args,
		}
		r.results = append(r.results, command)
	}

	return r
}

func (vow *Vow) runCmd(cmd *exec.Cmd, w io.Writer) (result cmdResult) {
	var b bytes.Buffer
	mw := io.MultiWriter(&b, w)

	cmd.Stdout = mw
	cmd.Stderr = mw

	cmd.Stdout.Write([]byte(strings.Join(cmd.Args, " ") + "\n"))
	if err := cmd.Start(); err != nil {
		mw.Write([]byte(err.Error() + "\n"))
		result.failed = true
		return
	}

	if err := cmd.Wait(); err != nil {
		result.failed = true
	}

	result.ps = cmd.ProcessState
	result.output = b.Bytes()
	result.command = cmd.Path
	result.args = cmd.Args[1:]

	return
}
