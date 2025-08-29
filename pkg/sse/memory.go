package sse

import (
	"context"
	"sync"
	"time"
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
	m.subsMu.Lock()
	defer m.subsMu.Unlock()

	// If the subscriber exists and time since last
	// sent event is greater than the timeStep (declared in sse.NewSubscriber)
	// then actually send the event
	// We could also use percentageStep if i find a way
	// to track that without breaking the upload 
	// process
	// if ev.Percentage - sub.lastPct > sub.pctStep 
	if sub, ok := m.subs[id]; ok &&
	(time.Since(sub.lastTime) >= sub.timeStep ||
	ev.Percentage - sub.lastPct >= 1 ||
	ev.Percentage == 100){
		//fmt.Printf("Last Time %v, was before %v\n", time.Since(sub.lastTime), sub.timeStep)
		//sub.lastPct = ev.Percentage
		sub.lastTime = time.Now()
		// Select is used here as non blocking operation
		select {
		case sub.Ch  <- *ev:
		default:
		}
	}
}
