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
)

func SendCommand(ip string) *cobra.Command {
	return &cobra.Command{
		Use:   "send file_path",
		Short: "Send a file",
		Long:  "Send a file to a computer",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, err := os.ReadFile(args[0])
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

			smsg := api.Message{
				Protocol:   api.INITIAL_CONNECT,
				Session_ID: session_id,
			}

			err = conn.WriteJSON(smsg)
			if err != nil {
				log.Fatal(err)
			}

			var A *pake.Pake

			for {
				var msg api.Message
				err := conn.ReadJSON(&msg)
				if err != nil {
					return
				}

				switch msg.Protocol {
				case api.CONFIRMATION:
					fmt.Println("Confirmed connection to HUB")
					// TODO: Failing because role == 0 for sender? Sender does not fail when role == 1
					A, err = pake.InitCurve([]byte(passphrase), 0, "siec")
					if err != nil {
						log.Fatal(err)
					}

					smsg := &api.Message{
						Protocol:   api.SHARE_PBK,
						Session_ID: session_id,
						PB_Key:     A.Bytes(),
					}
					conn.WriteJSON(smsg)

				case api.SHARE_PBK:
					fmt.Println("Received PBK")
					err = A.Update(msg.PB_Key)
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		},
	}
}
