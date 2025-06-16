package sse

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

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
		defer r.subsMu.RUnlock()
		sub, ok := r.subs[id]
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
	data, _ := json.Marshal(ev)
	fmt.Println(string(data))
	err := r.client.Publish(c, "upload:" + id, data).Err()
	if err != nil {
		fmt.Println(err)
	}
}
