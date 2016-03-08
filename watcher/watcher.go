package watcher

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	fsn "gopkg.in/fsnotify.v1"
)

type Watcher interface {
	Watch(string) error
}

type FSWatcher struct {
	fsn          *fsn.Watcher
	done         chan struct{}
	mtimes       map[string]time.Time
	watching     map[string]struct{}
	watchDir     string
	ignoredItems []string
	Event        chan struct{}
}

func New(ignoredItems []string) (*FSWatcher, error) {
	f, err := fsn.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FSWatcher{
		fsn:          f,
		ignoredItems: ignoredItems,
		mtimes:       map[string]time.Time{},
		done:         make(chan struct{}),
		watching:     map[string]struct{}{},
		Event:        make(chan struct{}),
	}, nil
}

func (w *FSWatcher) Watch(path string) error {
	w.watchDir = path
	// this can never return false since we will always
	// have at least one file in the directory (.snag.yml)
	_ = w.watch(path)
	w.Event <- struct{}{}

	for {
		select {
		case ev := <-w.fsn.Events:
			var queueBuild bool
			switch {
			case isCreate(ev.Op):
				queueBuild = w.watch(ev.Name)
			case isDelete(ev.Op):
				if _, ok := w.watching[ev.Name]; ok {
					w.fsn.Remove(ev.Name)
					delete(w.watching, ev.Name)
				}
				queueBuild = true
			case isModify(ev.Op):
				queueBuild = true
			}
			if queueBuild {
				w.maybeQueue(ev.Name)
			}
		case err := <-w.fsn.Errors:
			log.Println("error:", err)
		case <-w.done:
			return nil
		}
	}
}

func (w *FSWatcher) maybeQueue(path string) {
	if w.isExcluded(path) {
		return
	}

	stat, err := os.Stat(path)
	if err != nil {
		// we couldn't find the file
		// most likely a deletion
		delete(w.mtimes, path)
		w.Event <- struct{}{}
		return
	}

	mtime := stat.ModTime()
	lasttime := w.mtimes[path]
	if !mtime.Equal(lasttime) {
		// the file has been modified and the
		// file system event wasn't bogus
		w.mtimes[path] = mtime
		w.Event <- struct{}{}
	}
}

func (w *FSWatcher) watch(path string) bool {
	var shouldBuild bool
	if _, ok := w.watching[path]; ok {
		return false
	}
	filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
		if fi == nil {
			return filepath.SkipDir
		}

		if !fi.IsDir() {
			shouldBuild = true
			return nil
		}

		if w.isExcluded(p) {
			return filepath.SkipDir
		}

		if err := w.fsn.Add(p); err != nil {
			return err
		}
		w.watching[p] = struct{}{}

		return nil
	})
	return shouldBuild
}

func (w *FSWatcher) isExcluded(path string) bool {
	// get the relative path
	path = strings.TrimPrefix(path, w.watchDir+string(filepath.Separator))

	for _, p := range w.ignoredItems {
		if globMatch(p, path) {
			return true
		}
	}
	return false
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
