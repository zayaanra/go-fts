package fts

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/schollz/pake/v3"
	"github.com/sethvargo/go-diceware/diceware"
	"github.com/spf13/cobra"

	"github.com/zayaanra/go-fts/internal/crypt"
	"github.com/zayaanra/go-fts/pkg/api"
)

type Sender struct {
	HubAddress string
	conn *websocket.Conn
	FileData []byte
	SessionID string
	SessionKey []byte
	Passphrase string
	P *pake.Pake
}

func (s *Sender) Rendevous() error {
	p, err := pake.InitCurve([]byte(s.Passphrase), 0, "siec")
	if err != nil {
		return err
	}
	s.P = p

	conn, _, err := websocket.DefaultDialer.Dial("ws://" + s.HubAddress + ":8080/ws", nil)
	if err != nil {
		return err
	}
	s.conn = conn

	err = s.conn.WriteJSON(api.Message{
		Protocol: api.CONNECT,
		SessionID: s.SessionID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Sender) Listen() error {
	for {
		var msg api.Message
		err := s.conn.ReadJSON(&msg)
		if err != nil {
			return err
		}

		switch msg.Protocol {
		case api.ACKNOWLEDGE:
			fmt.Println("Confirmed connection to HUB")
			err = handleAcknowledgement(s.conn, s.SessionID, s.P.Bytes())

		case api.SHARE_PUBLIC_KEY:
			fmt.Println("Received PBK from B")
			sessionKey, err := s.HandlePublicKeyExchange(msg.PublicKey)
			if err != nil {
				return err
			}
			s.SessionKey = sessionKey

			ln, _ := net.Listen("tcp", ":0")
		
		// case api.SHARE_CONNECTION_INFO:
		// 	fmt.Println("Received connection info from B")
		// 	err = s.HandleIPExchange(msg.Data)
		}

		if err != nil {
			return err
		}
	}
}

func (s *Sender) HandlePublicKeyExchange(publicKey []byte) ([]byte, error) {
	err := s.P.Update(publicKey)
	if err != nil {
		return nil, err
	}
	
	sessionKey, err := s.P.SessionKey()
	if err != nil {
		return nil, err
	}

	encrypted, err := sec.EncryptAES([]byte(s.conn.LocalAddr().String()), sessionKey)
	if err != nil {
		return nil, err
	}

	err = s.conn.WriteJSON(api.Message{
		Protocol: api.SHARE_CONNECTION_INFO,
		SessionID: s.SessionID,
		Data: encrypted,
	})
	if err != nil {
		return nil, err
	}

	return sessionKey, nil
}

func (s *Sender) HandleIPExchange(data []byte) error {
	decrypted, err := sec.DecryptAES(data, s.SessionKey)
	if err != nil {
		return err
	}

	receiverIP := string(decrypted) + ":8090"
	log.Println(receiverIP)
	conn, err := net.Dial("tcp", receiverIP)
	if err != nil {
		log.Fatal(err)
	}

	fileData, err := json.Marshal(api.Message{
		Protocol: api.SHARE_FILE_DATA,
		Data: s.FileData,
	})

	encrypted, err := sec.EncryptAES(fileData, s.SessionKey)
	if err != nil {
		log.Fatal(err)
	}

	_, err = conn.Write(encrypted)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Println("Sent file to receiver")

	return nil
}

func (s *Sender) Close() error {
	if err := s.conn.Close(); err != nil {
		return err
	}
	return nil
}

func handleAcknowledgement(conn *websocket.Conn, sessionID string, publicKey []byte) error {
	err := conn.WriteJSON(api.Message{
		Protocol: api.SHARE_PUBLIC_KEY,
		SessionID: sessionID,
		PublicKey: publicKey,
	})
	return err
}

func SendCommand(ip string) *cobra.Command {
	return &cobra.Command{
		Use:   "send [file-path]",
		Short: "Send a file",
		Long:  "Send a file to a machine",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: What if file is too large to be stored in memory? (can stream data to receiver)
			data, err := os.ReadFile(args[0])
			if err != nil {
				log.Fatal(err)
			}
			
			// TODO: Return errors up the chain and handle them in the RunE field of the Cobra command
			// TODO: Ensure PAKE session key is passed through Key Derivation Function (KDF) like HKDF

			list, _ := diceware.Generate(5)
			session_id := uuid.New().String()[:5]
			passphrase := session_id + "-" + strings.Join(list, "-")
			
			var p api.Peer = &Sender{
				HubAddress: ip,
				FileData: data,
				SessionID: session_id,
				Passphrase: passphrase,
			}

			if err := p.Rendevous(); err != nil {
				log.Fatal(err)
			}
			defer p.Close()

			fmt.Println("On the receiving machine, run the receive command and enter the following code:")
			fmt.Println(passphrase)

			if err := p.Listen(); err != nil {
				log.Fatal(err)
			}
		},
	}
}
