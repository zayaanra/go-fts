package fts

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/schollz/pake/v3"
	"github.com/spf13/cobra"

	"github.com/zayaanra/go-fts/pkg/api"
	"github.com/zayaanra/go-fts/internal/sec"
)

type Receiver struct {
	HubAddress string
	conn *websocket.Conn
	FilePath string
	SessionID string
	SessionKey []byte
	Passphrase string
	P *pake.Pake
}

func (r *Receiver) Start() error {
	p, err := pake.InitCurve([]byte(r.Passphrase), 1, "siec")
	if err != nil {
		return err
	}
	r.P = p

	conn, _, err := websocket.DefaultDialer.Dial("ws://" + r.HubAddress + ":8080/ws", nil)
	if err != nil {
		return err
	}
	r.conn = conn

	err = r.conn.WriteJSON(api.Message{
		Protocol:   api.INITIAL_CONNECT,
		Session_ID: r.SessionID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r * Receiver) Listen() error {
	for {
		var msg api.Message
		err := r.conn.ReadJSON(&msg)
		if err != nil {
			return err
		}

		switch msg.Protocol {
		case api.CONFIRMATION:
			fmt.Println("Confirmed connection to HUB")

		case api.SEND_A_TO_B:
			fmt.Println("Received PBK from A")
			sessionKey, err := handleAtoB(r.P, r.conn, r.SessionID, msg.PB_Key)
			if err != nil {
				return err
			}
			r.SessionKey = sessionKey

		case api.SHARE_CONNECTION_INFO:
			fmt.Println("Received connection info from A")
			err = handleIPExchange_2(r.conn, r.SessionID, r.SessionKey)
			if err != nil {
				return err
			}
			// session_key, err = B.SessionKey()
			// if err != nil {
			// 	log.Fatal(err)
			// }

			// decrypted, err := sec.DecryptAES(msg.Data, session_key)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			
		case api.SEND_FILE_DATA:
			fmt.Println("Receiving file data from A")
			
			// decrypted, err := sec.DecryptAES(msg.Data, session_key)
			// if err != nil {
			// 	log.Fatal(err)
			// }

			// os.WriteFile(args[0], decrypted, 0777)
		}
	}
}

func handleAtoB(p *pake.Pake, conn *websocket.Conn, sessionID string, publicKey []byte) ([]byte, error) {
	err := p.Update(publicKey)
	if err != nil {
		return nil, err
	}

	err = conn.WriteJSON(api.Message{
		Protocol:   api.SEND_B_TO_A,
		Session_ID: sessionID,
		PB_Key:     p.Bytes(),
	})
	if err != nil {
		return nil, err
	}

	sessionKey, err := p.SessionKey()
	if err != nil {
		return nil, err
	}

	return sessionKey, nil
}

func handleIPExchange_2(conn *websocket.Conn, sessionID string, sessionKey []byte) error {
	encrypted, err := sec.EncryptAES([]byte(conn.LocalAddr().String()), sessionKey)
	if err != nil {
		return err
	}

	err = conn.WriteJSON(api.Message{
		Protocol: api.SHARE_CONNECTION_INFO,
		Session_ID: sessionID,
		Data: encrypted,
	})
	if err != nil {
		return err
	}

	return nil
}

func ReceiveCommand(ip string) *cobra.Command {
	return &cobra.Command{
		Use:   "receive file_path",
		Short: "Receive a file",
		Long:  "Receive a file from a machine",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Enter the code shared with you:")

			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()

			passphrase := scanner.Text()
			sessionID := strings.Split(passphrase, "-")[0]
			
			r := Receiver{
				HubAddress: ip,
				FilePath: args[0],
				SessionID: sessionID,
				Passphrase: passphrase,
			}
			err := r.Start()
			if err != nil {
				log.Fatal(err)
			}
			
			err = r.Listen()
			if err != nil {
				log.Fatal(err)
			}
		},
	}
}
