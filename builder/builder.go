package builder

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/Tonkpils/snag/exchange"
)

type Config struct {
	Build      []string
	Run        []string
	DepWarning string
	Verbose    bool
}

type Builder interface {
	Build(exchange.Event)
}

type CmdBuilder struct {
	ex         exchange.SendListener
	mtx        sync.RWMutex
	depWarning string
	buildCmds  [][]string
	runCmds    [][]string
	curVow     *vow

	verbose bool
}

func New(ex exchange.SendListener, c Config) Builder {
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

	b := &CmdBuilder{
		ex:         ex,
		buildCmds:  buildCmds,
		runCmds:    runCmds,
		depWarning: c.DepWarning,
		verbose:    c.Verbose,
	}

	ex.Listen("rebuild", b.Build)

	return b
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

func (b *CmdBuilder) stopCurVow() {
	b.mtx.Lock()
	if b.curVow != nil {
		b.curVow.Stop()
	}
	b.mtx.Unlock()
}

func (b *CmdBuilder) Build(_ exchange.Event) {
	b.stopCurVow()

	b.mtx.Lock()

	if len(b.depWarning) > 0 {
		fmt.Printf("Deprecation Warnings!\n%s", b.depWarning)
	}

	// setup the first command
	firstCmd := b.buildCmds[0]
	strs := []string{
		fmt.Sprintf("%s %s", firstCmd[0], strings.Join(firstCmd[1:], " ")),
	}
	b.curVow = VowTo(b.ex, firstCmd[0], firstCmd[1:]...)

	// setup the remaining commands
	for i := 1; i < len(b.buildCmds); i++ {
		cmd := b.buildCmds[i]
		strs = append(strs, fmt.Sprintf("%s %s", cmd[0], strings.Join(cmd[1:], " ")))
		b.curVow = b.curVow.Then(cmd[0], cmd[1:]...)
	}

	b.ex.Send("update-commands", strs)

	// setup all parallel commands
	for i := 0; i < len(b.runCmds); i++ {
		cmd := b.runCmds[i]
		b.curVow = b.curVow.ThenAsync(cmd[0], cmd[1:]...)
	}
	b.curVow.Verbose = b.verbose
	go b.curVow.Exec()

	b.mtx.Unlock()
}
