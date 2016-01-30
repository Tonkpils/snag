package main

import (
	"flag"
	"fmt"
)

var (
	cliCmds argSlice
	version bool
	verbose bool
)

func init() {
	flag.Var(&cliCmds, "c", "List of commands to execute")
	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.BoolVar(&version, "version", false, "display snag's version")
}

type argSlice []string

func (c *argSlice) String() string {
	return fmt.Sprintf("%s", *c)
}

func (a *argSlice) Set(value string) error {
	*a = append(*a, value)
	return nil
}
