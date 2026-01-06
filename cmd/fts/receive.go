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
			session_id := strings.Split(passphrase, "-")[0]

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
			
			B, err := pake.InitCurve([]byte(passphrase), 1, "siec")
			if err != nil {
				log.Fatal(err)
			}

			var session_key []byte

			for {
				var msg api.Message
				err := conn.ReadJSON(&msg)
				if err != nil {
					return
				}

				switch msg.Protocol {
				case api.CONFIRMATION:
					fmt.Println("Confirmed connection to HUB")

				case api.SEND_A_TO_B:
					fmt.Println("Received PBK")

					err = B.Update(msg.PB_Key)
					if err != nil {
						log.Fatal(err)
					}

					conn.WriteJSON(api.Message{
						Protocol:   api.SEND_B_TO_A,
						Session_ID: session_id,
						PB_Key:     B.Bytes(),
					})
				
				case api.SHARE_CONNECTION_INFO:
					fmt.Println("Received connection info from A")
					
					session_key, err = B.SessionKey()
					if err != nil {
						log.Fatal(err)
					}

					// decrypted, err := sec.DecryptAES(msg.Data, session_key)
					// if err != nil {
					// 	log.Fatal(err)
					// }
					
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

				case api.SEND_FILE_DATA:
					fmt.Println("Receiving file data from A")
					
					decrypted, err := sec.DecryptAES(msg.Data, session_key)
					if err != nil {
						log.Fatal(err)
					}

					os.WriteFile(args[0], decrypted, 0777)
				}
			}
		},
	}
}
