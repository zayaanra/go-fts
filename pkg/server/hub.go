package server

import (
	"encoding/json"
	"log"

	"github.com/zayaanra/go-fts/pkg/api"
)

type Hub struct {
	clients 	map[*Client]bool
	register 	chan *Client
	unregister 	chan *Client
	broadcast 	chan []byte

	rooms		map[string]*Room
}

type ExtendedClient struct {
	c *Client
	ready bool
	public_key []byte
}

type Room struct {
	a *ExtendedClient
	b *ExtendedClient
	exchanged bool
}

func NewHub() *Hub {
	return &Hub{
		clients: 	make(map[*Client]bool),
		register: 	make(chan *Client),
		unregister: make(chan *Client),
		broadcast:	make(chan []byte),
		rooms:		make(map[string]*Room),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub) FillRoom(client *Client, session_id string) {
	room, ok := h.rooms[session_id]
	if !ok {
		h.rooms[session_id] = &Room{
			a: &ExtendedClient{c: client, ready: false, public_key: nil},
			b: nil,
			exchanged: false,
		}
	} else {
		room.b = &ExtendedClient{c: client, ready: false, public_key: nil}

		smsg := api.Message{Protocol: api.CONFIRMATION}
		data, err := json.Marshal(smsg)
		if err != nil {
			log.Println(err)
			return
		}

		room.a.ready = true
		room.b.ready = true

		room.a.c.send <- data
		room.b.c.send <- data
	}
}

func (h *Hub) ExchangePBKs(client *Client, sessionID string, publicKey []byte) {
	room, ok := h.rooms[sessionID]
	if !ok {
		return
	}

	if room.exchanged {
		return
	}

	if !room.a.ready || !room.b.ready {
		return
	}

	if room.a == nil || room.b == nil {
		return
	}

	smsg, _ := json.Marshal(
		api.Message{
			Protocol: api.SHARE_PUBLIC_KEY,
			PublicKey: publicKey,
		},
	)

	switch client {
	case room.a.c:
		room.b.c.send <- smsg
	case room.b.c:
		room.a.c.send <- smsg
	}
}

func (h *Hub) ExchangeConnections(client *Client, session_id string, msg api.Message) {
	room, ok := h.rooms[session_id]
	if !ok {
		return
	}

	smsg, _ := json.Marshal(msg)
	switch client {
	case room.a.c:
		room.b.c.send <- smsg
	case room.b.c:
		room.a.c.send <- smsg
	}
}