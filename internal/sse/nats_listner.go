// internal/sse/nats_listener.go
package sse

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

func StartNATSListener(natsURL string, broker *Broker) error {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return err
	}
	js, err := nc.JetStream()
	if err != nil {
		return err
	}

	// Clean subscription
	_, err = js.Subscribe(
		"todo.events",
		func(m *nats.Msg) {
			log.Printf("RAW: %s", string(m.Data))

			var evt struct {
				UserID string `json:"user_id"`
			}
			if err := json.Unmarshal(m.Data, &evt); err != nil {
				log.Printf("invalid event: %v", err)
				_ = m.Ack()
				return
			}

			log.Printf("EVENT userID: %s", evt.UserID)
			broker.PublishTo(evt.UserID, string(m.Data))
			_ = m.Ack()
		},
		nats.Durable("sse-updates"),
		nats.ManualAck(),
	)
	if err != nil {
		return fmt.Errorf("subscribe error: %w", err)
	}

	log.Println("Subscribed to todo.events for SSE")
	return nil
}
