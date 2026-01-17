package server

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/zayaanra/go-fts/api"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 5012
)

type Client struct {
	mailbox *Mailbox
	conn 	*websocket.Conn
	send 	chan []byte
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
}

func (c *Client) readPump() {
	defer func() {
		c.mailbox.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		msg := api.Message{}
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			log.Println(err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		switch msg.Protocol {
		case api.CONNECT:
			c.mailbox.FillRoom(c, &msg)
		default:
			c.mailbox.ExchangeData(c, &msg)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWS(mailbox *Mailbox, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{mailbox: mailbox, conn: conn, send: make(chan []byte)}
	client.mailbox.register <- client

	go client.writePump()
	go client.readPump()
}
