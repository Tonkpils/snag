package exchange

import "sync"

type queue []func(Event)

type Listener interface {
	Listen(string, func(Event))
}

type Sender interface {
	Send(string, interface{})
}

type SendListener interface {
	Sender
	Listener
}

type Event struct {
	Name    string
	Payload interface{}
}

type Exchange struct {
	mtx    sync.RWMutex
	queues map[string]queue
}

func New() SendListener {
	return &Exchange{
		queues: map[string]queue{},
	}
}

func (ex *Exchange) Listen(event string, fn func(Event)) {
	ex.mtx.Lock()
	ex.queues[event] = append(ex.queues[event], fn)
	ex.mtx.Unlock()
}

func (ex *Exchange) Send(event string, data interface{}) {
	ex.mtx.RLock()
	for _, fn := range ex.queues[event] {
		go fn(Event{Name: event, Payload: data})
	}
	ex.mtx.RUnlock()
}
