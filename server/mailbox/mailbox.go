package server

import (
	"encoding/json"
	"log"

	"github.com/zayaanra/go-fts/api"
)

type Mailbox struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte

	rooms map[string]*Room
}

type Room struct {
	a *Client
	b *Client
}

func NewMailbox() *Mailbox {
	return &Mailbox{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		rooms:      make(map[string]*Room),
	}
}

func (m *Mailbox) Run() {
	for {
		select {
		case client := <-m.register:
			m.clients[client] = true
		case client := <-m.unregister:
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				close(client.send)
			}
		case message := <-m.broadcast:
			for client := range m.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(m.clients, client)
				}
			}
		}
	}
}

func (m *Mailbox) FillRoom(client *Client, msg *api.Message) {
	sessionID := msg.SessionID

	room, ok := m.rooms[sessionID]
	if !ok {
		m.rooms[sessionID] = &Room{
			a: client,
			b: nil,
		}
	} else {
		room.b = client

		msg := &api.Message{Protocol: api.ACKNOWLEDGE}
		data, err := json.Marshal(msg)
		if err != nil {
			log.Println(err)
			return
		}

		room.a.send <- data
		room.b.send <- data
	}
}

func (m *Mailbox) ExchangeData(client *Client, msg *api.Message) {
	sessionID := msg.SessionID
	room, ok := m.rooms[sessionID]
	if !ok {
		return
	}

	if room.a == nil || room.b == nil {
		return
	}

	smsg, _ := json.Marshal(msg)
	switch client {
	case room.a:
		room.b.send <- smsg
	case room.b:
		room.a.send <- smsg
	}	
}