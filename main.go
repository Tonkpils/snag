package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	buildTool string
	packages  string
	buildArgs args
	vetArgs   args
	testArgs  args
)

const (
	buildToolDesc = "build tool used to run commands. (Godeps, GB and Go are the only ones currently supported)"
	buildDesc     = "comma delimited list of arguments given to the build command"
	vetDesc       = "comma delimited list of arguments given to the vet command"
	testDesc      = "comma delimited list of arguments given to the test command"
	packageDesc   = "comma delimited list of packages to run commands on"
)

type args []string

func (a *args) String() string {
	return fmt.Sprint(*a)
}

func (a *args) Set(value string) error {
	if value != "" {
		for _, v := range strings.Split(value, ",") {
			*a = append(*a, v)
		}
	}

	return nil
}

func init() {
	flag.StringVar(&packages, "packages", "./...", packageDesc)
	flag.StringVar(&buildTool, "build-tool", "go", buildToolDesc)
	flag.Var(&buildArgs, "build", buildDesc)
	flag.Var(&vetArgs, "vet", vetDesc)
	flag.Var(&testArgs, "test", testDesc)
}

func main() {
	flag.Parse()

	b, err := NewBuilder(
		strings.Split(packages, ","),
		buildArgs,
		vetArgs,
		testArgs,
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
