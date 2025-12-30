package fts

import (
	"io"
	"net/http"
	"os"
	"log"

	"github.com/spf13/cobra"
)

func SendCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "send ip_address file_path",
		Short: "Send a file",
		Long:  "Send a file to a computer",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			target := "http://" + args[0] + ":" + "3333"
			
			data, err := os.Open(args[1])
			if err != nil {
				log.Fatal(err)
			}

			req, err := http.NewRequest("POST", target, data)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}

			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Status Code: %d\n", resp.StatusCode)
			log.Printf("Response Body: %s\n", body)
		},
	}
}
