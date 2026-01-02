package fts

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/zayaanra/go-fts/pkg/api"

	"github.com/google/uuid"
	"github.com/sethvargo/go-diceware/diceware"
	"github.com/spf13/cobra"
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
			passphrase := strings.Join(list, "-")

			id := uuid.New().String()[:5]
			id += "-" + passphrase 

			fmt.Fprintln(os.Stdout, "On the receiving machine, run the receive command and enter the following code:")
			fmt.Println(id)

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

			sig := make(chan os.Signal, 1)
			signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

			<-sig

		},
	}
}