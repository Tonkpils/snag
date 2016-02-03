package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	cliCmds argSlice
	version bool
	verbose bool
)

func init() {
	flag.Var(&cliCmds, "c", "List of commands to execute")
	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.BoolVar(&version, "version", false, "[DEPRECATED: use 'snag version'] display snag's version")

	flag.Usage = func() {
		usage := `Usage of %s:
    %s [COMMAND]

Commands:

    init    	Generate a snag file %q used for configuration and execution
    version 	Display snag's version

Flags:
`
		fmt.Fprintf(os.Stderr, usage, os.Args[0], os.Args[0], SnagFile)
		flag.PrintDefaults()
	}
}

type argSlice []string

func (c *argSlice) String() string {
	return fmt.Sprintf("%s", *c)
}

func (a *argSlice) Set(value string) error {
	*a = append(*a, value)
	return nil
}
