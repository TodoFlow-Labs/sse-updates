// internal/sse/nats_listener.go
package sse

import (
	"github.com/nats-io/nats.go"
	"log"
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
		broker.Publish(string(m.Data))
		_ = m.Ack()
	}, nats.Durable("sse-updates"), nats.ManualAck())
	if err != nil {
		return err
	}

	log.Println("Subscribed to todo.events for SSE")
	return nil
}
