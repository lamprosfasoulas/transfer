package sse

import (
	"context"
	"time"
)

type Dispatcher interface {
	AddSubscriber(context.Context, string, *Subscriber)
	DelSubscriber(context.Context, string)
	SendEvent(context.Context, string, *ProgressEvent)
}

//var (
//	Subs = make(map[string]*Subscriber)
//	SubsMu sync.RWMutex
//	//Sessions = make(map[string]bool)
//)

// ProgressEvent is a struct used to capture upload detail.
// It is then send with SendEventToSubscriber.
type ProgressEvent struct {
    Filename    string  `json:"filename,omitempty"`
    Bytes       int64   `json:"bytes"`
    TotalBytes  int64   `json:"total_bytes,omitempty"`
    Percentage  int `json:"percentage"`
    Message     string  `json:"message,omitempty"`
}

// Subscriber is each new connection to the SSEHandler.
type Subscriber struct {
	Ch 		chan ProgressEvent // Channel capturing ProgressEvents
	Closing chan struct{} // Channel to signal the closing of the connection
	lastPct int
	lastTime time.Time
	pctStep int
	timeStep time.Duration
}

// NewProgressEvent is a constructor for ProgressEvent
func NewProgressEvent(filename, message string, bytes, totalbytes int64, pct int) *ProgressEvent{
	return &ProgressEvent{
		Filename: filename,
		Bytes: bytes,
		TotalBytes: totalbytes,
		Percentage: pct,
		Message: message,
	}
}

// NewSubscriber is a constructor for each new Subscriber.
func NewSubscriber (pStep int, tStep time.Duration) *Subscriber {
	return &Subscriber{
		Ch: make(chan ProgressEvent),
		Closing: make(chan struct {}),
		lastTime: time.Now().Add(-500 * time.Millisecond),
		pctStep: pStep,
		timeStep: tStep,
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



// SSEHandler is the http handler for streaming server events,
// (file upload status) to the front end client.

// SendEventToSubscriber is used to write new event
// to each subscriber's event channel
