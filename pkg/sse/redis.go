package sse

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

type RedisDispatcher struct {
	subs map[string]*Subscriber
	subsMu sync.RWMutex
	client *redis.Client
	ctx context.Context
}

func NewRedisDispatcher(addr string) *RedisDispatcher {
	r := &RedisDispatcher{
		subs: make(map[string]*Subscriber),
		ctx: context.Background(),
		client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
	// Listens to the upload:* channel on redis
	go r.listen()

	return r
}

func (r *RedisDispatcher) listen() {
	pubsub := r.client.PSubscribe(r.ctx, "upload:*")
	ch := pubsub.Channel()

	for msg := range ch {
		var id string
		var ev ProgressEvent

		strs := strings.Split(msg.Channel,":")
		if len(strs) > 1 {
			id = strs[1]
		}

		json.Unmarshal([]byte(msg.Payload), &ev)

		r.subsMu.RLock()
		sub, ok := r.subs[id]
		r.subsMu.RUnlock()
		if ok {
			select {
			case sub.Ch <- ev:
			default:
			}
		}
	}
}

func (r *RedisDispatcher) AddSubscriber(c context.Context, id string, sub *Subscriber) {
	r.subsMu.Lock()
	defer r.subsMu.Unlock()
	r.subs[id] = sub
}

func (r *RedisDispatcher) DelSubscriber(c context.Context, id string) {
	r.subsMu.Lock()
	defer r.subsMu.Unlock()
	delete(r.subs, id)
}

func (r *RedisDispatcher) SendEvent(c context.Context, id string, ev *ProgressEvent) {
	r.subsMu.Lock()
	defer r.subsMu.Unlock()
	// This Message is a copy from pkg/sse/memory.go
	// If the subscriber exists and time since last
	// sent event is greater than the timeStep (declared in sse.NewSubscriber)
	// then actually send the event
	// We could also use percentageStep if i find a way
	// to track that without breaking the upload 
	// process
	// if ev.Percentage - sub.lastPct > sub.pctStep 
	if sub, ok := r.subs[id]; ok && 
	(time.Since(sub.lastTime) >= sub.timeStep || ev.Percentage == 100) {
		//fmt.Printf("Last Time %v, was before %v\n", time.Since(sub.lastTime), sub.timeStep)
		//sub.lastPct = ev.Percentage
		sub.lastTime = time.Now()
		data, _ := json.Marshal(ev)
		err := r.client.Publish(c, "upload:" + id, data).Err()
		if err != nil {
			fmt.Println(err)
		}
	}
}
