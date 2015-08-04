package main

import (
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type config struct {
	Script           []string `yaml:"script"`
	ExcludeDirectory []string `yaml:"exclude_directory"`
	Verbose          bool     `yaml:"verbose"`
}

func main() {
	in, err := ioutil.ReadFile(".snag.yml")
	if err != nil {
		log.Fatal("Could not find '.snag.yml' in your current directory")
	}

	var c config
	if err := yaml.Unmarshal(in, &c); err != nil {
		log.Fatalf("Could not parse yml file. %s\n", err)
	}

	if len(c.Script) == 0 {
		log.Fatal("You must have at least 1 command in your '.snag.yml'")
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
