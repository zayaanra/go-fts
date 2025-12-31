package fts

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/zayaanra/go-fts/pkg/api"

	"github.com/spf13/cobra"
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

			// TODO: Generate passphrase and session ID to send to relay server

			conn, err := net.Dial("tcp", fmt.Sprintf(ip + ":8090"))
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()

			message := api.Message{
				Session_ID: "test",
				Data: data,
			}

			bytes, err := json.Marshal(message)
			if err != nil {
				log.Fatal(err)
			}

			_, err = conn.Write(bytes)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
}
