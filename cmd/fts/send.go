package fts

import (
	"fmt"
	"log"
	"os"

	// "os/signal"
	"strings"
	// "syscall"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/schollz/pake/v3"
	"github.com/sethvargo/go-diceware/diceware"
	"github.com/spf13/cobra"

	"github.com/zayaanra/go-fts/pkg/api"
	"github.com/zayaanra/go-fts/internal/sec"
)

func SendCommand(ip string) *cobra.Command {
	return &cobra.Command{
		Use:   "send file_path",
		Short: "Send a file",
		Long:  "Send a file to a computer",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			data, err := os.ReadFile(args[0])
			if err != nil {
				log.Fatal(err)
			}

			list, err := diceware.Generate(5)

			session_id := uuid.New().String()[:5]
			passphrase := session_id + "-" + strings.Join(list, "-")

			fmt.Println("On the receiving machine, run the receive command and enter the following code:")
			fmt.Println(passphrase)

			conn, _, err := websocket.DefaultDialer.Dial("ws://" + ip + ":8080/ws", nil)
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()

			err = conn.WriteJSON(api.Message{
				Protocol:   api.INITIAL_CONNECT,
				Session_ID: session_id,
			})

			if err != nil {
				log.Fatal(err)
			}

			A, err := pake.InitCurve([]byte(passphrase), 0, "siec")
			if err != nil {
				log.Fatal(err)
			}

			var session_key []byte

			for {
				var msg api.Message
				err := conn.ReadJSON(&msg)
				if err != nil {
					log.Fatal(err)
				}

				switch msg.Protocol {
				case api.CONFIRMATION:
					fmt.Println("Confirmed connection to HUB")

					conn.WriteJSON(api.Message{
						Protocol: api.SEND_A_TO_B,
						Session_ID: session_id,
						PB_Key: A.Bytes(),
					})

				case api.SEND_B_TO_A:
					fmt.Println("Received PBK from B")
					
					err = A.Update(msg.PB_Key)
					if err != nil {
						log.Fatal(err)
					}
					
					session_key, err = A.SessionKey()
					if err != nil {
						log.Fatal(err)
					}

					myIP := conn.LocalAddr().String()
					encrypted, err := sec.EncryptAES([]byte(myIP), session_key)
					if err != nil {
						log.Fatal(err)
					}

					conn.WriteJSON(api.Message{
						Protocol: api.SHARE_CONNECTION_INFO,
						Session_ID: session_id,
						Data: encrypted,
					})
				
				case api.SHARE_CONNECTION_INFO:
					fmt.Println("Received connection info from B")

					_, err := sec.DecryptAES(msg.Data, session_key)
					if err != nil {
						log.Fatal(err)
					}

					encrypted, err := sec.EncryptAES(data, session_key)
					if err != nil {
						log.Fatal(encrypted)
					}

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

				}
			}
		},
	}
}
