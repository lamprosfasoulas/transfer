package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lamprosfasoulas/transfer/pkg/sse"
)

func (m *MainHandlers) SSEHandler (w http.ResponseWriter, r *http.Request) {
	uploadID := r.URL.Query().Get("id")
	if uploadID == "" {
		http.Error(w, "Missing id", http.StatusBadRequest)
		return
	}

	sub := sse.NewSubscriber()
	m.Dispatcher.AddSubscriber(r.Context(), uploadID, sub)
	defer m.Dispatcher.DelSubscriber(r.Context(), uploadID)
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

	fmt.Fprintf(w, "event: connected\ndata: %s\n\n", `{"message": "Connected, ready for upload"}`)
	flusher.Flush()

	defer func() {
		m.Dispatcher.DelSubscriber(r.Context(), uploadID)
		close(sub.Ch)
	}()

	for {
		select {
		case ev := <-sub.Ch:
			data, _ := json.Marshal(ev)
			fmt.Fprintf(w, "data: %s\n\n", data)
			//fmt.Println(data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		case <-sub.Closing:
			return
		}
	}
}
