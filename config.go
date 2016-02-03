package main

import (
	"errors"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type config struct {
	DepWarnning  string
	Script       []string `yaml:"script"`
	Build        []string `yaml:"build"`
	Run          []string `yaml:"run"`
	IgnoredItems []string `yaml:"ignore"`
	Verbose      bool     `yaml:"verbose"`
}

func parseConfig() (config, error) {
	var c config

	// if we have any cliCmds, set them to our build phase
	c.Build = cliCmds

	// if build phase is still empty try and find the snag.yml file
	if len(c.Build) == 0 {
		in, err := ioutil.ReadFile(SnagFile)
		if err != nil {
			return c, fmt.Errorf("could not find %q in your current directory", SnagFile)
		}

		if err := yaml.Unmarshal(in, &c); err != nil {
			return c, fmt.Errorf("could not parse snag file: %s\n", err)
		}
	}

	// if both script and build are specified
	// blow up and tell the user to use build
	if len(c.Script) != 0 && len(c.Build) != 0 {
		return c, errors.New("cannot use 'script' and 'build' together. The 'script' tag is deprecated, please use 'build' instead.")
	}

	// if script has something, tell the user it's deprecated
	// and set whatever its contents are to build
	if len(c.Script) != 0 {
		c.DepWarnning += "*\tThe use of 'script' in the yaml file has been deprecated and will be removed in the future.\n\tPlease start using 'build' instead.\n\n"
		c.Build = c.Script
	}

	if len(c.Build) == 0 {
		return c, errors.New("you must specify at least 1 command.")
	}

	c.Verbose = verbose || c.Verbose
	return c, nil
}
