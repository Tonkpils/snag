package exchange

import "sync"

type queue []func(interface{})

type SendListener interface {
	Listen(string, func(interface{}))
	Send(string, interface{})
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

func (ex *Exchange) Listen(event string, fn func(interface{})) {
	ex.mtx.Lock()
	ex.queues[event] = append(ex.queues[event], fn)
	ex.mtx.Unlock()
}

func (ex *Exchange) Send(event string, data interface{}) {
	ex.mtx.RLock()
	for _, fn := range ex.queues[event] {
		fn(data)
	}
	ex.mtx.RUnlock()
}
