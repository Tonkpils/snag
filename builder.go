package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Tonkpils/snag/vow"
	fsn "gopkg.in/fsnotify.v1"
)

var (
	mtimes = map[string]time.Time{}
)

// supported file extensions
const (
	GoExt = ".go"
)

// BuildTool represents a program that builds/tests code
type BuildTool string

// supported build tools
const (
	BuildToolGo    = "go"
	BuildToolGodep = "godep"
	BuildToolGB    = "gb"
)

type Bob struct {
	w        *fsn.Watcher
	mtx      sync.RWMutex
	curVow   *vow.Vow
	done     chan struct{}
	watching map[string]struct{}

	buildTool string
	buildArgs []string
	vetArgs   []string
	testArgs  []string
}

func NewBuilder(packages, build, vet, test []string) (*Bob, error) {
	w, err := fsn.NewWatcher()
	if err != nil {
		return nil, err
	}

	b := append([]string{"build"}, build...)
	b = append(b, packages...)

	v := append([]string{"vet"}, vet...)
	v = append(v, packages...)

	t := append([]string{"test"}, test...)
	t = append(t, packages...)

	return &Bob{
		w:         w,
		done:      make(chan struct{}),
		watching:  map[string]struct{}{},
		buildArgs: b,
		vetArgs:   v,
		testArgs:  t,
	}, nil
}

func (b *Bob) BuildWith(bt BuildTool) {
	b.buildTool = string(bt)
	if b.buildTool == BuildToolGodep {
		b.buildArgs = append([]string{"go"}, b.buildArgs...)
		b.testArgs = append([]string{"go"}, b.testArgs...)
	}
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

func (b *Bob) stopCurVow() {
	if b.curVow != nil {
		b.mtx.Lock()
		b.curVow.Stop()
		b.mtx.Unlock()
	}
}

func (b *Bob) execute() {
	b.stopCurVow()

	b.clearBuffer()
	b.mtx.Lock()
	b.curVow = vow.To(b.buildTool, b.buildArgs...).
		Then("go", b.vetArgs...).
		Then(b.buildTool, b.testArgs...)
	go b.curVow.Exec(os.Stdout)
	b.mtx.Unlock()
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
