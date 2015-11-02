package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type argSlice []string

func (c *argSlice) String() string {
	return fmt.Sprintf("%s", *c)
}

func (a *argSlice) Set(value string) error {
	*a = append(*a, value)
	return nil
}

type config struct {
	Script       []string `yaml:"script"`
	IgnoredItems []string `yaml:"ignore"`
	Verbose      bool     `yaml:"verbose"`
}

const (
	Version       = "1.1.1"
	VersionOutput = "Snag version " + Version
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

func main() {
	flag.Parse()
	if version {
		fmt.Println(VersionOutput)
		return
	}

	var c config
	if len(cliCmds) > 0 {
		c.Script = cliCmds
	} else {
		in, err := ioutil.ReadFile(".snag.yml")
		if err != nil {
			log.Fatal("Could not find '.snag.yml' in your current directory")
		}

		if err := yaml.Unmarshal(in, &c); err != nil {
			log.Fatalf("Could not parse yml file. %s\n", err)
		}

	}

	if len(c.Script) == 0 {
		log.Fatal("You must specify at least 1 command.")
	}

	if verbose {
		c.Verbose = verbose
	}

	b, err := NewBuilder(c)
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	b.Watch(wd)
}
