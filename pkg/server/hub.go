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

type Room struct {
	a	*Client
	b	*Client
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
	// If room exists, connect client B and broadcast to room that both clients (A and B) have been accepted
	if ok {
		room.b = client
		smsg := api.Message{Protocol: api.CONFIRMATION}
		data, err := json.Marshal(smsg)
		if err != nil {
			log.Println(err)
			return
		}
		room.a.send <- data
		room.b.send <- data
	// If room does not exist yet, create one with only client A
	} else {
		h.rooms[session_id] = &Room{a: client, b: nil}
	}
}

func (h *Hub) ExchangePBKs(client *Client, session_id string, pb_key []byte) {
	room, ok := h.rooms[session_id]
	if ok {
		smsg := api.Message{
			Protocol: api.SHARE_PBK,
			PB_Key: pb_key,
		}
		data, err := json.Marshal(smsg)
		if err != nil {
			log.Println(err)
			return
		}
		if room.a == client {
			room.b.send <- data
		} else {
			room.a.send <- data
		}
	}
}