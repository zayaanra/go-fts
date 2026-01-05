package fts

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/schollz/pake/v3"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"

	"github.com/zayaanra/go-fts/pkg/api"
)

func ReceiveCommand(ip string) *cobra.Command {
	return &cobra.Command{
		Use:   "receive file_path",
		Short: "Receive a file",
		Long:  "Receive a file from a machine",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			scanner := bufio.NewScanner(os.Stdin)
			fmt.Println("Enter the code shared with you:")
			scanner.Scan()
			passphrase := scanner.Text()

			session_id := strings.Split(passphrase, "-")[0]

			conn, _, err := websocket.DefaultDialer.Dial("ws://" + ip + ":8080/ws", nil)
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()

			message := api.Message{
				Protocol: api.INITIAL_CONNECT,
				Session_ID: session_id,
			}

			err = conn.WriteJSON(message)
			if err != nil {
				log.Fatal(err)
			}
			
			var B *pake.Pake

			for {
				var msg api.Message
				err := conn.ReadJSON(&msg)
				if err != nil {
					return
				}

				switch msg.Protocol {
				case api.CONFIRMATION:
					fmt.Println("Confirmed connection to HUB")
					B, err = pake.InitCurve([]byte(passphrase), 1, "siec")
					if err != nil {
						log.Fatal(err)
					}
					
					smsg := &api.Message{
						Protocol: api.SHARE_PBK,
						Session_ID: session_id,
						PB_Key: B.Bytes(),
					}
					conn.WriteJSON(smsg)
				
				case api.SHARE_PBK:
					fmt.Println("Received PBK")
					err = B.Update(msg.PB_Key)
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		},
	}
}
