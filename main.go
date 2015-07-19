package main

import (
	"flag"
	"log"
	"os"
)

var (
	buildArgs string
	vetArgs   string
	testArgs  string
	packages  string
)

const (
	buildDesc   = "arguments given to the build command"
	vetDesc     = "arguments given to the vet command"
	testDesc    = "arguments given to the test command"
	packageDesc = "packages to run commands on"
)

func init() {
	flag.StringVar(&packages, "packages", "./...", packageDesc)
	flag.StringVar(&buildArgs, "build", "", buildDesc)
	flag.StringVar(&vetArgs, "vet", "", vetDesc)
	flag.StringVar(&testArgs, "test", "", testDesc)
}

func main() {
	flag.Parse()

	b, err := NewBuilder(packages, buildArgs, vetArgs, testArgs)
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
