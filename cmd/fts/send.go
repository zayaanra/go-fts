package fts

import (
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/schollz/pake/v3"
	"github.com/sethvargo/go-diceware/diceware"
	"github.com/spf13/cobra"

	"github.com/zayaanra/go-fts/pkg/api"
	"github.com/zayaanra/go-fts/internal/sec"
)

type Sender struct {
	HubAddress string
	conn *websocket.Conn
	FilePath string
	SessionID string
	SessionKey []byte
	Passphrase string
	P *pake.Pake
}

func (s *Sender) Start() error {
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
		Protocol:   api.INITIAL_CONNECT,
		Session_ID: s.SessionID,
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
		case api.CONFIRMATION:
			fmt.Println("Confirmed connection to HUB")
			err = handleConfirmation(s.conn, s.SessionID, s.P.Bytes())

		case api.SEND_B_TO_A:
			fmt.Println("Received PBK from B")
			sessionKey, err := handleBtoA(s.P, s.conn, s.SessionID, msg.PB_Key)
			if err != nil {
				return err
			}
			s.SessionKey = sessionKey
		
		case api.SHARE_CONNECTION_INFO:
			fmt.Println("Received connection info from B")
			err = handleIPExchange(s.conn, s.SessionID, s.SessionKey, msg.Data)
		}

		if err != nil {
			return err
		}
	}
}

func handleConfirmation(conn *websocket.Conn, sessionID string, publicKey []byte) error {
	err := conn.WriteJSON(api.Message{
		Protocol: api.SEND_A_TO_B,
		Session_ID: sessionID,
		PB_Key: publicKey,
	})
	return err
}

func handleBtoA(p *pake.Pake, conn *websocket.Conn, sessionID string, publicKey []byte) ([]byte, error) {
	err := p.Update(publicKey)
	if err != nil {
		return nil, err
	}
	
	sessionKey, err := p.SessionKey()
	if err != nil {
		return nil, err
	}

	encrypted, err := sec.EncryptAES([]byte(conn.LocalAddr().String()), sessionKey)
	if err != nil {
		return nil, err
	}

	err = conn.WriteJSON(api.Message{
		Protocol: api.SHARE_CONNECTION_INFO,
		Session_ID: sessionID,
		Data: encrypted,
	})
	if err != nil {
		return nil, err
	}

	return sessionKey, nil
}

func handleIPExchange(conn *websocket.Conn, sessionID string, sessionKey []byte, data []byte) error {
	_, err := sec.DecryptAES(data, sessionKey)
	if err != nil {
		return err
	}

	// encrypted, err := sec.EncryptAES(data, sessionKey)
	// if err != nil {
	// 	return err
	// }

	// receiver_ip := string(decrypted)
	
	// smsg := &api.Message{
	// 	Protocol: api.SEND_FILE_DATA,
	// 	Session_ID: session_id,
	// 	Data: encrypted,
	// }

	// TODO: Open direct TCP connection to B to send file data to

	// ip := strings.Split(string(B_ip), ":")[0]

	// listener, err := net.L
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer newConn.Close()


	// newConn.WriteJSON(smsg)
	return nil
}

func SendCommand(ip string) *cobra.Command {
	return &cobra.Command{
		Use:   "send file_path",
		Short: "Send a file",
		Long:  "Send a file to a computer",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: What if file is too large to be stored in memory?
			// data, err := os.ReadFile(args[0])
			// if err != nil {
			// 	log.Fatal(err)
			// }
			
			// TODO: Return errors up the chain and handle them in the RunE field of the Cobra command
			// TODO: Ensure PAKE session key is passed through Key Derivation Function (KDF) like HKDF

			list, err := diceware.Generate(5)
			session_id := uuid.New().String()[:5]
			passphrase := session_id + "-" + strings.Join(list, "-")

			s := Sender{
				HubAddress: ip, 
				FilePath: args[0], 
				SessionID: session_id, 
				Passphrase: passphrase,
			}

			err = s.Start()
			if err != nil {
				log.Fatal(err)
			}
			defer s.conn.Close()

			fmt.Println("On the receiving machine, run the receive command and enter the following code:")
			fmt.Println(passphrase)

			err = s.Listen()
			if err != nil {
				log.Fatal(err)
			}
		},
	}
}
