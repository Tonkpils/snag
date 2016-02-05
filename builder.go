package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Tonkpils/snag/vow"
	"github.com/rjeczalik/notify"
	"github.com/shiena/ansicolor"
)

var mtimes = map[string]time.Time{}
var clearBuffer = func() {
	fmt.Print("\033c")
}

type Bob struct {
	mtx      sync.RWMutex
	curVow   *vow.Vow
	done     chan struct{}
	watchDir string

	depWarning   string
	buildCmds    [][]string
	runCmds      [][]string
	ignoredItems []string

	verbose bool
}

func NewBuilder(c config) (*Bob, error) {
	parseCmd := func(cmd string) []string {
		c := strings.Split(cmd, " ")

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
		done:         make(chan struct{}),
		buildCmds:    buildCmds,
		runCmds:      runCmds,
		depWarning:   c.DepWarnning,
		ignoredItems: c.IgnoredItems,
		verbose:      c.Verbose,
	}, nil
}

func replaceEnv(cmds []string) {
	for i, c := range cmds {
		if !strings.HasPrefix(c, "$$") {
			continue
		}

		cmds[i] = os.Getenv(strings.TrimPrefix(c, "$$"))
	}
}

func (b *Bob) Close() {
	close(b.done)
}

func (b *Bob) Watch(path string) error {
	b.watchDir = path

	b.execute()

	// Make the channel buffered to ensure no event is dropped. Notify will drop
	// an event if the receiver is not able to keep up the sending pace.
	c := make(chan notify.EventInfo, 1)

	// Set up a watchpoint listening on events within current working directory.
	// Dispatch all events to c
	if err := notify.Watch(".", c, notify.All); err != nil {
		return err
	}
	defer notify.Stop(c)

	// Block until an event is received or builder is closed
	for {
		select {
		case ev := <-c:
			if b.shouldQueue(ev.Path()) {
				b.execute()
			}
		case <-b.done:
			return nil
		}
	}
}

func (b *Bob) shouldQueue(path string) bool {
	if b.isExcluded(path) {
		return false
	}

	stat, err := os.Stat(path)
	if err != nil {
		// we couldn't find the file
		// most likely a deletion
		delete(mtimes, path)
		return true
	}

	mtime := stat.ModTime()
	lasttime := mtimes[path]
	if !mtime.Equal(lasttime) {
		// the file has been modified and the
		// file system event wasn't bogus
		mtimes[path] = mtime
		return true
	}

	return false
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
