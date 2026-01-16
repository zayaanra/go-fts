package peer

import (
	"github.com/gorilla/websocket"
	"github.com/schollz/pake/v3"
	"github.com/zayaanra/go-fts/pkg/api"
)

const (
	PAKE_INITIATOR = 0
	PAKE_RESPONDER = 1
)

type Peer struct {
	Conn       *websocket.Conn
	Curve      *pake.Pake
	FileData   []byte
	Role       int
	SenderIP   string
	ReceiverIP string
	Session    *Session
}

type Session struct {
	ID         string
	Passphrase string
	Key        []byte
}

func NewPeer(role int, sessionID string, passphrase string) *Peer {
	s := &Session{ID: sessionID, Passphrase: passphrase}
	return &Peer{Role: role, Session: s}
}

func (p *Peer) Rendevous(mailboxIP string) error {
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+mailboxIP+":8080/ws", nil)
	if err != nil {
		return err
	}
	p.Conn = conn

	err = p.Conn.WriteJSON(api.Message{
		Protocol:  api.CONNECT,
		SessionID: p.Session.ID,
	})
	if err != nil {
		return err
	}

	curve, err := pake.InitCurve([]byte(p.Session.Passphrase), p.Role, "siec")
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
			err = handleAck(p)
			if err != nil {
				return err
			}

		case api.SHARE_PUBLIC_KEY:
			err = handleSharePublicKey(p, msg.PublicKey)
			if p.Role == PAKE_RESPONDER {
				return err
			}

		case api.SHARE_IP:
			err = handleShareIP(p, msg.Data)
			if p.Role == PAKE_INITIATOR {
				return err
			}
		}
	}
}

func (p *Peer) Close() error {
	err := p.Conn.Close()
	if err != nil {
		return err
	}
	return nil
}
