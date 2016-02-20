package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Tonkpils/snag/vow"
	"github.com/shiena/ansicolor"
	fsn "gopkg.in/fsnotify.v1"
)

var mtimes = map[string]time.Time{}
var clearBuffer = func() {
	fmt.Print("\033c")
}

type Bob struct {
	w        *fsn.Watcher
	mtx      sync.RWMutex
	curVow   *vow.Vow
	done     chan struct{}
	watching map[string]struct{}
	watchDir string

	depWarning   string
	buildCmds    [][]string
	runCmds      [][]string
	ignoredItems []string

	verbose bool
}

func NewBuilder(c config) (*Bob, error) {
	w, err := fsn.NewWatcher()
	if err != nil {
		return nil, err
	}

	parseCmd := func(cmd string) (c []string) {
		s := bufio.NewScanner(strings.NewReader(cmd))
		s.Split(splitFunc)
		for s.Scan() {
			c = append(c, s.Text())
		}

		// check for environment variables inside script
		if strings.Contains(cmd, "$$") {
			replaceEnv(c)
		}
		return c
	}

	buildCmds := make([][]string, len(c.Build))
	for i, s := range c.Build {
		buildCmds[i] = parseCmd(s)
	}

	runCmds := make([][]string, len(c.Run))
	for i, s := range c.Run {
		runCmds[i] = parseCmd(s)
	}

	return &Bob{
		w:            w,
		done:         make(chan struct{}),
		watching:     map[string]struct{}{},
		buildCmds:    buildCmds,
		runCmds:      runCmds,
		depWarning:   c.DepWarnning,
		ignoredItems: c.IgnoredItems,
		verbose:      c.Verbose,
	}, nil
}

func splitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	advance, token, err = bufio.ScanWords(data, atEOF)
	if err != nil {
		return
	}

	if len(token) == 0 {
		return
	}

	b := token[0]
	if b != '"' && b != '\'' {
		return
	}

	if token[len(token)-1] == b {
		return
	}

	chunk := data[advance-1:]
	i := bytes.IndexByte(chunk, b)
	if i == -1 {
		advance = len(data)
		token = append(token, chunk...)
		return
	}

	advance += i
	token = append(token, chunk[:i+1]...)

	return
}

func replaceEnv(cmds []string) {
	for i, c := range cmds {
		if !strings.HasPrefix(c, "$$") {
			continue
		}

		cmds[i] = os.Getenv(strings.TrimPrefix(c, "$$"))
	}
}

func (b *Bob) Close() error {
	close(b.done)
	return b.w.Close()
}

func (b *Bob) Watch(path string) error {
	b.watchDir = path
	// this can never return false since we will always
	// have at least one file in the directory (.snag.yml)
	_ = b.watch(path)
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
	if b.isExcluded(path) {
		return
	}

	stat, err := os.Stat(path)
	if err != nil {
		// we couldn't find the file
		// most likely a deletion
		delete(mtimes, path)
		b.execute()
		return
	}

	mtime := stat.ModTime()
	lasttime := mtimes[path]
	if !mtime.Equal(lasttime) {
		// the file has been modified and the
		// file system event wasn't bogus
		mtimes[path] = mtime
		b.execute()
	}
}

func (b *Bob) stopCurVow() {
	b.mtx.Lock()
	if b.curVow != nil {
		b.curVow.Stop()
	}
	b.mtx.Unlock()
}

func (b *Bob) execute() {
	b.stopCurVow()

	clearBuffer()
	b.mtx.Lock()

	if len(b.depWarning) > 0 {
		fmt.Printf("Deprecation Warnings!\n%s", b.depWarning)
	}

	// setup the first command
	firstCmd := b.buildCmds[0]
	b.curVow = vow.To(firstCmd[0], firstCmd[1:]...)

	// setup the remaining commands
	for i := 1; i < len(b.buildCmds); i++ {
		cmd := b.buildCmds[i]
		b.curVow = b.curVow.Then(cmd[0], cmd[1:]...)
	}

	// setup all parallel commands
	for i := 0; i < len(b.runCmds); i++ {
		cmd := b.runCmds[i]
		b.curVow = b.curVow.ThenAsync(cmd[0], cmd[1:]...)
	}
	b.curVow.Verbose = b.verbose
	go b.curVow.Exec(ansicolor.NewAnsiColorWriter(os.Stdout))

	b.mtx.Unlock()
}

func (b *Bob) watch(path string) bool {
	var shouldBuild bool
	if _, ok := b.watching[path]; ok {
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

		if b.isExcluded(p) {
			return filepath.SkipDir
		}

		if err := b.w.Add(p); err != nil {
			return err
		}
		b.watching[p] = struct{}{}

		return nil
	})
	return shouldBuild
}

func (b *Bob) isExcluded(path string) bool {
	// get the relative path
	path = strings.TrimPrefix(path, b.watchDir+string(filepath.Separator))

	for _, p := range b.ignoredItems {
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
