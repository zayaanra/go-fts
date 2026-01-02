package fts

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

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
			text := scanner.Text()

			id := strings.Split(text, "-")[0]

			conn, _, err := websocket.DefaultDialer.Dial("ws://" + ip + ":8080/ws", nil)
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()

			message := api.Message{
				Protocol: api.INITIAL_CONNECT,
				Session_ID: id,
			}

			err = conn.WriteJSON(message)
			if err != nil {
				log.Fatal(err)
			}

			// message := api.Message{
			// 	Protocol: api.INITIAL_CONNECT,
			// 	Session_ID: id,
			// }
			
		},
	}
}
