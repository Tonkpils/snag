package main

import (
	"flag"
	"log"
	"os"
)

const (
	Version       = "1.1.1"
	VersionOutput = "Snag version " + Version
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)
}

func main() {
	flag.Parse()
	if version {
		log.Println(VersionOutput)
		return
	}

	c, err := parseConfig()
	if err != nil {
		log.Fatal(err)
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
