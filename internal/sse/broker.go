// internal/sse/broker.go
package sse

import (
	"sync"
)

type Broker struct {
	clients    map[string]map[chan string]bool // userID -> clients
	register   chan clientReg
	unregister chan clientReg
	broadcast  chan userMsg
	lock       sync.RWMutex
}

type clientReg struct {
	userID string
	ch     chan string
}

type userMsg struct {
	userID string
	msg    string
}

func NewBroker() *Broker {
	return &Broker{
		clients:    make(map[string]map[chan string]bool),
		register:   make(chan clientReg),
		unregister: make(chan clientReg),
		broadcast:  make(chan userMsg),
	}
}

func (b *Broker) Run() {
	for {
		select {
		case reg := <-b.register:
			b.lock.Lock()
			if b.clients[reg.userID] == nil {
				b.clients[reg.userID] = make(map[chan string]bool)
			}
			b.clients[reg.userID][reg.ch] = true
			b.lock.Unlock()

		case reg := <-b.unregister:
			b.lock.Lock()
			if clients, ok := b.clients[reg.userID]; ok {
				if _, exists := clients[reg.ch]; exists {
					delete(clients, reg.ch)
					close(reg.ch)
					if len(clients) == 0 {
						delete(b.clients, reg.userID)
					}
				}
			}
			b.lock.Unlock()

		case msg := <-b.broadcast:
			b.lock.RLock()
			for ch := range b.clients[msg.userID] {
				select {
				case ch <- msg.msg:
				default:
				}
			}
			b.lock.RUnlock()
		}
	}
}

func (b *Broker) PublishTo(userID, msg string) {
	b.broadcast <- userMsg{userID, msg}
}

func (b *Broker) Subscribe(userID string) chan string {
	ch := make(chan string, 10)
	b.register <- clientReg{userID, ch}
	return ch
}

func (b *Broker) Unsubscribe(userID string, ch chan string) {
	b.unregister <- clientReg{userID, ch}
}
