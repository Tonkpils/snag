// +build windows

package main

import (
	"os"
	"os/exec"
)

func init() {
	clearBuffer = func() {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			panic(err)
		}
	}
}
