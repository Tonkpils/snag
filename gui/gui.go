package gui

import (
	"fmt"
	"sync"

	"github.com/Tonkpils/snag/exchange"
	"github.com/gizak/termui"
)

var stateColors = map[string]string{
	"running": "fg-yellow",
	"failed":  "fg-red",
	"passed":  "fg-green",
}

type GUI struct {
	CommandsView  *termui.List
	commandsMutex *sync.Mutex
	exchange      exchange.Listener
	stateColors   map[string]string
}

func New(ex exchange.Listener) (*GUI, error) {
	if err := termui.Init(); err != nil {
		return nil, err
	}

	termui.Handle("/sys/kbd/q", func(termui.Event) {
		termui.StopLoop()
	})

	termui.Handle("/sys/kbd/C-x", func(termui.Event) {
		termui.StopLoop()
	})

	ls := termui.NewList()
	ls.ItemFgColor = termui.ColorRed
	ls.BorderLabel = "Snag"
	ls.Height = 10
	ls.Width = 100
	ls.Y = 0

	ui := &GUI{
		CommandsView:  ls,
		exchange:      ex,
		commandsMutex: &sync.Mutex{},
	}

	ex.Listen("update-commands", ui.Handle)
	ex.Listen("update-command", ui.Handle)

	return ui, nil
}

func (g *GUI) Close() {
	termui.Close()
}

func (g *GUI) Loop() {
	termui.Loop()
}

func (g *GUI) Handle(ev exchange.Event) {
	switch ev.Name {
	case "update-command":
		payload := ev.Payload.(map[string]interface{})
		index := payload["index"].(int)
		command := payload["command"].(string)
		state := payload["state"].(string)
		g.commandsMutex.Lock()
		g.UpdateCommand(index, state, command)
		g.commandsMutex.Unlock()
	case "update-commands":
		g.commandsMutex.Lock()
		g.UpdateCommandList(ev.Payload.([]string))
		g.commandsMutex.Unlock()
	default:
		panic(ev.Name)
	}
}

func (g *GUI) UpdateCommand(index int, state, cmd string) {
	g.CommandsView.Items[index] = fmt.Sprintf("[[%d] %s](%s)", index, cmd, stateColors[state])
	termui.Render(g.CommandsView)
}

func (g *GUI) UpdateCommandList(cmds []string) {
	items := make([]string, len(cmds))
	for idx, cmd := range cmds {
		items[idx] = fmt.Sprintf("[%d] %s", idx, cmd)
	}

	g.CommandsView.Items = items
	termui.Render(g.CommandsView)
}
