package vow

import (
	"bytes"
	"log"
	"os"
	"os/exec"
)

type Vow struct {
	cmds   []*exec.Cmd
	errlog *log.Logger
}

func To(name string, args ...string) *Vow {
	cmd := exec.Command(name, args...)
	return &Vow{
		cmds:   []*exec.Cmd{cmd},
		errlog: log.New(os.Stderr, "To - ", log.Lshortfile),
	}
}

func (vow *Vow) Then(name string, args ...string) *Vow {
	vow.cmds = append(vow.cmds, exec.Command(name, args...))
	return vow
}

func (vow *Vow) Exec() *Result {
	r := new(Result)
	var runCount int
	for runCount < len(vow.cmds) {
		command := vow.runCmd(vow.cmds[runCount])
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

func (vow *Vow) runCmd(cmd *exec.Cmd) (result cmdResult) {
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b

	if err := cmd.Start(); err != nil {
		vow.errlog.Println("err start", err)
		// could not start cmd
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
