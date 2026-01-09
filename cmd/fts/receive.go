package fts

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/schollz/pake/v3"
	"github.com/spf13/cobra"

	"github.com/zayaanra/go-fts/internal/sec"
	"github.com/zayaanra/go-fts/pkg/api"
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
		SessionID: r.SessionID,
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

		case api.SHARE_PUBLIC_KEY:
			fmt.Println("Received PBK from A")
			sessionKey, err := r.HandlePublicKeyExchange(msg.PublicKey)
			if err != nil {
				return err
			}
			r.SessionKey = sessionKey

		case api.SHARE_CONNECTION_INFO:
			fmt.Println("Received connection info from A")
			err = r.HandleIPExchange(nil)
			if err != nil {
				return err
			}

			// TODO: better way to do this net.Listen?
			ln, err := net.Listen("tcp", ":8090")
			if err != nil {
				log.Fatal(err)
			}
			defer ln.Close()

			conn, err := ln.Accept()
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()

			encrypted, err := io.ReadAll(conn)
			if err != nil {
				log.Fatal(err)
			}

			decrypted, err := sec.DecryptAES(encrypted, r.SessionKey)
			if err != nil {
				log.Fatal(err)
			}

			var msg api.Message
			if err := json.Unmarshal(decrypted, &msg); err != nil {
				log.Fatal(err)
			}

			// TODO: Write file to local filepath
			// log.Printf("Received message: %+v\n", string(msg.Data))
			
			
		case api.SHARE_FILE_DATA:
			fmt.Println("Receiving file data from A")
			
			decrypted, err := sec.DecryptAES(msg.Data, r.SessionKey)
			if err != nil {
				log.Fatal(err)
			}

			os.WriteFile(r.FilePath, decrypted, 0777)
		}
	}
}

func (r *Receiver) HandlePublicKeyExchange(publicKey []byte) ([]byte, error) {
	err := r.P.Update(publicKey)
	if err != nil {
		return nil, err
	}

	err = r.conn.WriteJSON(api.Message{
		Protocol:   api.SHARE_PUBLIC_KEY,
		SessionID: r.SessionID,
		PublicKey:     r.P.Bytes(),
	})
	if err != nil {
		return nil, err
	}

	sessionKey, err := r.P.SessionKey()
	if err != nil {
		return nil, err
	}

	return sessionKey, nil
}

func (r *Receiver) HandleIPExchange(data []byte) error {
	host := strings.Split(r.conn.LocalAddr().String(), ":")[0]
	encrypted, err := sec.EncryptAES([]byte(host), r.SessionKey)
	if err != nil {
		return err
	}

	err = r.conn.WriteJSON(api.Message{
		Protocol: api.SHARE_CONNECTION_INFO,
		SessionID: r.SessionID,
		Data: encrypted,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *Receiver) Close() error {
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

func ReceiveCommand(ip string) *cobra.Command {
	return &cobra.Command{
		Use:   "receive [file-path]",
		Short: "Receive a file",
		Long:  "Receive a file from a machine",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Enter the code shared with you:")

			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()

			passphrase := scanner.Text()
			sessionID := strings.Split(passphrase, "-")[0]
			
			var p api.Peer = &Receiver{
				HubAddress: ip,
				FilePath: args[0],
				SessionID: sessionID,
				Passphrase: passphrase,
			}

			if err := p.Start(); err != nil {
				log.Fatal(err)
			}
			defer p.Close()

			
			if err := p.Listen(); err != nil {
				log.Fatal(err)
			}
		},
	}
}
