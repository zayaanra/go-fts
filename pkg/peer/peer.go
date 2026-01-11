package peer

import (
	"fmt"
	"net"

	"github.com/gorilla/websocket"
	"github.com/schollz/pake/v3"
	"github.com/zayaanra/go-fts/pkg/api"
)

const (
    PAKE_INITIATOR = 0
    PAKE_RESPONDER = 1
)

type Peer struct {
	Role int

	Conn *websocket.Conn

	SessionID string
	SessionKey []byte
	Passphrase string

	Curve *pake.Pake
}

func NewPeer(role int, sessionID string, passphrase string) *Peer {
	return &Peer{Role: role, SessionID: sessionID, Passphrase: passphrase}
}

func (p *Peer) Rendevous(mailboxIP string) error {
	conn, _, err := websocket.DefaultDialer.Dial("ws://" + mailboxIP + ":8080/ws", nil)
	if err != nil {
		return err
	}
	p.Conn = conn

	err = p.Conn.WriteJSON(api.Message{
		Protocol: api.CONNECT,
		SessionID: p.SessionID,
	})
	if err != nil {
		return err
	}

	curve, err := pake.InitCurve([]byte(p.Passphrase), p.Role, "siec")
	if err != nil {
		return err
	}
	p.Curve = curve

	return nil
}

func (p *Peer) ListenWS() error {
	for {
		var msg api.Message
		err := p.Conn.ReadJSON(&msg)
		if err != nil {
			return err
		}

		switch msg.Protocol {
		case api.ACKNOWLEDGE:
			fmt.Println("Received ACK from Mailbox Server")
			if p.Role == PAKE_INITIATOR {
				err = p.Conn.WriteJSON(api.Message{
					Protocol: api.SHARE_PUBLIC_KEY,
					SessionID: p.SessionID,
					PublicKey: p.Curve.Bytes(),
				})
				if err != nil {
					return err
				}
			}

		case api.SHARE_PUBLIC_KEY:
			fmt.Println("Received Public Key")
			err = p.Curve.Update(msg.PublicKey)
			if err != nil {
				return err
			}

			if p.Role == PAKE_RESPONDER {
				err = p.Conn.WriteJSON(api.Message{
					Protocol: api.SHARE_PUBLIC_KEY,
					SessionID: p.SessionID,
					PublicKey: p.Curve.Bytes(),
				})
				if err != nil {
					return err
				}
			}

			sessionKey, _ := p.Curve.SessionKey()
			p.SessionKey = sessionKey
		}
	}
}

func (p *Peer) ListenTCP() error {
	if (p.Role == PAKE_RESPONDER) {
		ln, err := net.Listen("tcp", ":0")
		if err != nil {
			return err
		}

		// addr := ln.Addr().(*net.TCPAddr)


	}
	return nil
}

func (p *Peer) Close() error {
	err := p.Conn.Close()
	if err != nil {
		return err
	}
	return nil
}