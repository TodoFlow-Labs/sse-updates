// internal/sse/handler.go
package sse

import (
	"fmt"
	"log"
	"net/http"
)

type Handler struct {
	broker *Broker
}

func NewHandler(b *Broker) http.Handler {
	return &Handler{broker: b}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	val := r.Context().Value("user_id")
	userID, ok := val.(string)
	if !ok || userID == "" {
		log.Println("Missing or invalid user ID in context")
		http.Error(w, "X-User-ID required", http.StatusUnauthorized)
		return
	}

	log.Printf("SSE handler started for userID: %s", userID)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

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
