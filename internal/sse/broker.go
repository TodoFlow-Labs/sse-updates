// internal/sse/broker.go
package sse

import (
	"sync"
)

type Broker struct {
	clients    map[chan string]bool
	register   chan chan string
	unregister chan chan string
	broadcast  chan string
	lock       sync.RWMutex
}

func NewBroker() *Broker {
	return &Broker{
		clients:    make(map[chan string]bool),
		register:   make(chan chan string),
		unregister: make(chan chan string),
		broadcast:  make(chan string),
	}
}

func (b *Broker) Run() {
	for {
		select {
		case client := <-b.register:
			b.lock.Lock()
			b.clients[client] = true
			b.lock.Unlock()

		case client := <-b.unregister:
			b.lock.Lock()
			if _, ok := b.clients[client]; ok {
				delete(b.clients, client)
				close(client)
			}
			b.lock.Unlock()

		case msg := <-b.broadcast:
			b.lock.RLock()
			for client := range b.clients {
				select {
				case client <- msg:
				default:
				}
			}
			b.lock.RUnlock()
		}
	}
}

func (b *Broker) Publish(msg string) {
	b.broadcast <- msg
}

func (b *Broker) Subscribe() chan string {
	ch := make(chan string, 10)
	b.register <- ch
	return ch
}

func (b *Broker) Unsubscribe(ch chan string) {
	b.unregister <- ch
}
