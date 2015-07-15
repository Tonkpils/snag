package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	fsn "gopkg.in/fsnotify.v1"
)

var (
	mtimes   = map[string]time.Time{}
	watching = map[string]struct{}{}
)

func main() {
	w, err := fsn.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case ev := <-w.Events:
				fmt.Println(ev)
				var queueBuild bool
				switch {
				case isCreate(ev.Op):
					fmt.Println("Creating")
					queueBuild = watch(w, ev.Name)
				case isDelete(ev.Op):
					fmt.Println("Deleteing")
					if _, ok := watching[ev.Name]; ok {
						w.Remove(ev.Name)
						delete(watching, ev.Name)
					}

					queueBuild = true
				case isModify(ev.Op):
					queueBuild = true
				}
				if queueBuild {
					maybeQueue(ev.Name)
				}
			case err := <-w.Errors:
				log.Println("error:", err)
			}
		}
	}()

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	watch(w, wd)

	<-done
}

func watch(w *fsn.Watcher, path string) bool {
	var shouldBuild bool
	if _, ok := watching[path]; ok {
		return false
	}
	filepath.Walk(path, func(p string, fileInfo os.FileInfo, err error) error {
		if fileInfo == nil {
			return err
		}
		if fileInfo.IsDir() {
			if err := w.Add(p); err != nil {
				return err
			} else {
				watching[p] = struct{}{}
			}
		} else {
			shouldBuild = true
		}
		return nil
	})
	return shouldBuild
}

const (
	GoExt = ".go"
)

func maybeQueue(path string) {
	fmt.Println(filepath.Ext(path) == GoExt)
	if filepath.Ext(path) != GoExt {
		return
	}

	stat, err := os.Stat(path)
	if err == nil {
		mtime := stat.ModTime()
		lasttime := mtimes[path]
		if !mtime.Equal(lasttime) {
			mtimes[path] = mtime
			cmd := exec.Command("go", "test")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stdout

			if err := cmd.Run(); err != nil {
				fmt.Println(err)
			}
		}
	} else {
		log.Println(err)
		delete(mtimes, path)
		cmd := exec.Command("go", "test")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout

		if err := cmd.Run(); err != nil {
			fmt.Println(err)
		}
	}
}

func isCreate(op fsn.Op) bool {
	return op&fsn.Create == fsn.Create
}

func isDelete(op fsn.Op) bool {
	return op&fsn.Remove == fsn.Remove
}

func isModify(op fsn.Op) bool {
	return op&fsn.Write == fsn.Write ||
		op&fsn.Rename == fsn.Rename
}
