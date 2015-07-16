package main

import "log"

func main() {
	b, err := NewBuilder()
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	b.Watch()
}
