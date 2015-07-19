package vow

import "os"

// Result represents the outcome of the Vow that was run
type Result struct {
	// Failed will be true if any command was not successful
	Failed bool

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
