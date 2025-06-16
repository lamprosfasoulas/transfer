package sse

import (
	"context"
	"sync"
)


type MemDispatcher struct {
	subs map[string]*Subscriber
	subsMu sync.RWMutex
}

func NewMemDispatcher() *MemDispatcher {
	return &MemDispatcher{
		subs: make(map[string]*Subscriber),
	}
}

func (m *MemDispatcher) AddSubscriber(c context.Context, id string, sub *Subscriber) {
	m.subsMu.Lock()
	defer m.subsMu.Unlock()
	m.subs[id] = sub
}

func (m *MemDispatcher) DelSubscriber(c context.Context, id string) {
	m.subsMu.Lock()
	defer m.subsMu.Unlock()
	delete(m.subs, id)
}

func (m *MemDispatcher) SendEvent(c context.Context, id string, ev *ProgressEvent) {
	m.subsMu.RLock()
	defer m.subsMu.RUnlock()

	if sub, ok := m.subs[id]; ok {
		select {
		case sub.Ch  <- *ev:
		default:
		}
	}

}
