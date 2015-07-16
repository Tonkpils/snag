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
	mtimes = map[string]time.Time{}
)

const (
	GoExt = ".go"
)

type Bob struct {
	w        *fsn.Watcher
	done     chan struct{}
	watching map[string]struct{}
}

func NewBuilder() (*Bob, error) {
	w, err := fsn.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Bob{
		w:        w,
		done:     make(chan struct{}),
		watching: map[string]struct{}{},
	}, nil
}

func (b *Bob) Close() {
	b.w.Close()
	b.done <- struct{}{}
}

func (b *Bob) Watch() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case ev := <-b.w.Events:
				fmt.Println(ev)
				var queueBuild bool
				switch {
				case isCreate(ev.Op):
					fmt.Println("Creating")
					queueBuild = b.watch(ev.Name)
				case isDelete(ev.Op):
					fmt.Println("Deleteing")
					if _, ok := b.watching[ev.Name]; ok {
						b.w.Remove(ev.Name)
						delete(b.watching, ev.Name)
					}

					queueBuild = true
				case isModify(ev.Op):
					queueBuild = true
				}
				if queueBuild {
					b.maybeQueue(ev.Name)
				}
			case err := <-b.w.Errors:
				log.Println("error:", err)
			}
		}
	}()
	b.watch(wd)

	<-b.done

	return nil
}

func (b *Bob) maybeQueue(path string) {
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

func (b *Bob) watch(path string) bool {
	var shouldBuild bool
	if _, ok := b.watching[path]; ok {
		return false
	}
	filepath.Walk(path, func(p string, fileInfo os.FileInfo, err error) error {
		if fileInfo == nil {
			return err
		}
		if fileInfo.IsDir() {
			if err := b.w.Add(p); err != nil {
				return err
			} else {
				b.watching[p] = struct{}{}
			}
		} else {
			shouldBuild = true
		}
		return nil
	})
	return shouldBuild

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
