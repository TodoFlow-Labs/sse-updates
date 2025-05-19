// internal/sse/nats_listener.go
package sse

import (
	"encoding/json"
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

	_, err = js.Subscribe("todo.events", func(m *nats.Msg) {
		var evt struct {
			UserID string `json:"user_id"`
		}
		if err := json.Unmarshal(m.Data, &evt); err != nil {
			log.Printf("invalid event: %v", err)
			_ = m.Ack()
			return
		}

		broker.PublishTo(evt.UserID, string(m.Data))
		_ = m.Ack()
	}, nats.Durable("sse-updates"), nats.ManualAck())

	log.Println("Subscribed to todo.events for SSE")
	return nil
}
