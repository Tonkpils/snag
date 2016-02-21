package main

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/gizak/termui"
)

const (
	Version       = "1.2.0"
	VersionOutput = "Snag version " + Version
)

const SnagFile = ".snag.yml"

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)
}

func main() {
	flag.Parse()
	if flag.NArg() > 0 {
		if err := handleSubCommand(flag.Arg(0)); err != nil {
			log.Fatal(err)
		}
		return
	}

	if version {
		log.Println("The 'version' flag is deprecated. Use 'snag version'")
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

	if err := termui.Init(); err != nil {
		log.Fatal(err)
	}
	defer termui.Close()

	go b.Watch(wd)

	termui.Handle("/sys/kbd/q", func(termui.Event) {
		termui.StopLoop()
	})

	termui.Handle("/sys/kbd/C-x", func(termui.Event) {
		termui.StopLoop()
	})

	termui.Loop()
}

func handleSubCommand(cmd string) error {
	switch flag.Arg(0) {
	case "init":
		return initSnag()
	case "version":
		log.Println(VersionOutput)
		return nil
	default:
		flag.Usage()
		return nil
	}
}

func initSnag() error {
	if _, err := os.Stat(SnagFile); err == nil {
		return errors.New("snag file already exists")
	}

	f, err := os.Create(SnagFile)
	if err != nil {
		return err
	}
	defer f.Close()

	tmpl := `---
# Snag configuartion
#
# Make sure you modify this file to get started.
# If you have any questions please refer to https://github.com/Tonkpils/snag
#
# Verbose controls whether the process will output a command's output.
# verbose: true
#
# Use the ignore section to ignore files or directors from being watched.
# You can use 'gitignore' patterns for each item in the list.
# ignore:
#   - .git
#
# Build executes a list of commands sequentially
# build:
#   - echo 'Hello world'
`
	_, err = f.Write([]byte(tmpl))
	if err != nil {
		return err
	}

	success := `Successfully created sample configuration %q in your current directory.
Make sure you modify this file and run 'snag' to get going!`
	log.Printf(success, SnagFile)

	return nil
}
