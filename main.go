package main

import (
	"flag"
	"log"
	"os"
	"strings"
)

var (
	buildTool string
	buildArgs string
	vetArgs   string
	testArgs  string
	packages  string
)

const (
	buildToolDesc = "build tool used to run commands. (Godeps, GB and Go are the only ones currently supported)"
	buildDesc     = "comma delimited list of arguments given to the build command"
	vetDesc       = "comma delimited list of arguments given to the vet command"
	testDesc      = "comma delimited list of arguments given to the test command"
	packageDesc   = "comma delimited list of packages to run commands on"
)

func init() {
	flag.StringVar(&packages, "packages", "./...", packageDesc)
	flag.StringVar(&buildTool, "build-tool", "go", buildToolDesc)
	flag.StringVar(&buildArgs, "build", "", buildDesc)
	flag.StringVar(&vetArgs, "vet", "", vetDesc)
	flag.StringVar(&testArgs, "test", "", testDesc)
}

func main() {
	flag.Parse()

	b, err := NewBuilder(
		strings.Split(packages, ","),
		strings.Split(buildArgs, ","),
		strings.Split(vetArgs, ","),
		strings.Split(testArgs, ","),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	switch strings.ToLower(buildTool) {
	case "godep":
		b.BuildWith(BuildToolGodep)
	case "gb":
		b.BuildWith(BuildToolGB)
	default:
		b.BuildWith(BuildToolGo)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	b.Watch(wd)
}
