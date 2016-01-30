package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const (
	Version       = "1.1.1"
	VersionOutput = "Snag version " + Version
)

func main() {
	flag.Parse()
	if version {
		fmt.Println(VersionOutput)
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
