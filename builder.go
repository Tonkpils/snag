package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Tonkpils/snag/vow"
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

	packages  string
	buildArgs string
	vetArgs   string
	testArgs  string
}

func NewBuilder(packages, build, vet, test string) (*Bob, error) {
	w, err := fsn.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Bob{
		w:         w,
		done:      make(chan struct{}),
		watching:  map[string]struct{}{},
		packages:  packages,
		buildArgs: build,
		vetArgs:   vet,
		testArgs:  test,
	}, nil
}

func (b *Bob) Close() {
	b.w.Close()
	close(b.done)
}

func (b *Bob) Watch(path string) error {
	b.watch(path)
	b.execute()

	for {
		select {
		case ev := <-b.w.Events:
			var queueBuild bool
			switch {
			case isCreate(ev.Op):
				queueBuild = b.watch(ev.Name)
			case isDelete(ev.Op):
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
		case <-b.done:
			return nil
		}
	}
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
			b.execute()
		}
	} else {
		delete(mtimes, path)
		b.execute()
	}
}

func (b *Bob) execute() {
	b.clearBuffer()
	vow.To("go", "build", b.packages, b.buildArgs).
		Then("go", "vet", b.packages, b.vetArgs).
		Then("go", "test", b.testArgs, b.packages).
		Exec(os.Stdout)
}

func (b *Bob) clearBuffer() {
	fmt.Print("\033c")
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
			}
			b.watching[p] = struct{}{}
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
