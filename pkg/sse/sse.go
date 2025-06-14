package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)


var (
	Subs = make(map[string]*Subscriber)
	SubsMu sync.RWMutex
	//Sessions = make(map[string]bool)
)

// ProgressEvent is a struct used to capture upload detail.
// It is then send with SendEventToSubscriber.
type ProgressEvent struct {
    Filename    string  `json:"filename,omitempty"`
    Bytes       int64   `json:"bytes"`
    TotalBytes  int64   `json:"total_bytes,omitempty"`
    Percentage  float64 `json:"percentage"`
    Message     string  `json:"message,omitempty"`
}

// NewProgressEvent is a constructor for ProgressEvent
func NewProgressEvent(filename, message string, bytes, totalbytes int64, pct float64) *ProgressEvent{
	return &ProgressEvent{
		Filename: filename,
		Bytes: bytes,
		TotalBytes: totalbytes,
		Percentage: pct,
		Message: message,
	}
}

//Progress Reader (io.Reader) wrapper
//type progressReader struct {
//    src        io.Reader
//    total      int64 // total bytes to read (if known)
//    read       int64
//    filename   string
//	uploadID   string
//}

// Subscriber is each new connection to the SSEHandler.
type Subscriber struct {
	ch 		chan ProgressEvent // Channel capturing ProgressEvents
	closing chan struct{} // Channel to signal the closing of the connection
}

// NewSubscriber is a constructor for each new Subscriber.
func NewSubscriber () *Subscriber {
	return &Subscriber{
		ch: make(chan ProgressEvent),
		closing: make(chan struct {}),
	}
}

// SSEHandler is the http handler for streaming server events,
// (file upload status) to the front end client.
func SSEHandler (w http.ResponseWriter, r *http.Request) {
	uploadID := r.URL.Query().Get("id")
	if uploadID == "" {
		http.Error(w, "Missing id", http.StatusBadRequest)
		return
	}
	//SubsMu.RLock()
	//sub := Subs[uploadID]
	//SubsMu.RUnlock()
	//SSE Headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
	}

	sub := NewSubscriber()

	SubsMu.Lock()
	Subs[uploadID] = sub
	SubsMu.Unlock()

	fmt.Fprintf(w, "event: connected\ndata: %s\n\n", `{"message": "Connected, ready for upload"}`)
	flusher.Flush()

	defer func() {
		SubsMu.Lock()
		delete(Subs, uploadID)
		SubsMu.Unlock()
		close(sub.ch)
	}()

	for {
		select {
		case ev := <-sub.ch:
			data, _ := json.Marshal(ev)
			fmt.Fprintf(w, "data: %s\n\n", data)
			//fmt.Println(data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		case <-sub.closing:
			return
		}
	}
}

// SendEventToSubscriber is used to write new event
// to each subscriber's event channel
func SendEventToSubscriber(id string, ev *ProgressEvent) {
	SubsMu.RLock()
	defer SubsMu.RUnlock()
	if sub, ok := Subs[id]; ok {
		select {
		case sub.ch <- *ev:
		default:
		}
	}
}
