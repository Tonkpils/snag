package main

import (
	"log"
	"os"
)

func main() {
	b, err := NewBuilder()
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
