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

	userID := fmt.Sprint(r.Context().Value("user_id"))
	ch := h.broker.Subscribe(userID)
	defer h.broker.Unsubscribe(userID, ch)

	ctx := r.Context()
	for {
		select {
		case msg := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-ctx.Done():
			return
		}
	}
}
