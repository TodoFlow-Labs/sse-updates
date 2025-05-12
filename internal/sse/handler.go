// internal/sse/handler.go
package sse

import (
	"fmt"
	"net/http"
)

type Handler struct {
	broker *Broker
}

func NewHandler(b *Broker) http.Handler {
	return &Handler{broker: b}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	ch := h.broker.Subscribe()
	defer h.broker.Unsubscribe(ch)

	notify := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify
		h.broker.Unsubscribe(ch)
	}()

	for {
		select {
		case msg := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
