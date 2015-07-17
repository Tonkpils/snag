package vow

import (
	"fmt"
	"os"
)

// Result represents the outcome of all the command that
// were executed
type Result struct {
	// Whether or not the result was successful
	Failed   bool
	results  []cmdResult
	executed int
}

type cmdResult struct {
	command string
	args    []string

	ps     *os.ProcessState
	output []byte
	failed bool
}

func (r *Result) String() string {
	s := "To Result:\n"
	s += fmt.Sprintf("\tRan %d of %d commands\n", r.executed, len(r.results))

	for i := 0; i < len(r.results); i++ {
		cmdrslt := r.results[i]

		var outcome string
		switch {
		case i < r.executed-1 ||
			(i == r.executed-1 && !cmdrslt.failed):
			outcome = "SUCCESS - "
		case i >= r.executed:
			outcome = "SKIPPED - "
		case i == r.executed-1:
			outcome = "FAILED  - "
		}

		s += fmt.Sprintf("\t%s%s - %v\n", outcome, cmdrslt.command, cmdrslt.args)

		if cmdrslt.failed {
			s += fmt.Sprintf("Output:\n%s\n", cmdrslt.output)
		}
	}
	return s
}
