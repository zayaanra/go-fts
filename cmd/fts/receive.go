package fts

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"log"

	"github.com/spf13/cobra"
)

func ReceiveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "receive file_path",
		Short: "Receive a file",
		Long:  "Receive a file from a machine",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					log.Fatal(err)
				}
				err = os.WriteFile(args[0], body, 0777)
				if err != nil {
					log.Fatal(err)
				}
			})
			if err := http.ListenAndServe(":3333", mux); err != nil {
				if !errors.Is(err, http.ErrServerClosed) {
					fmt.Printf("error running http server: %s\n", err)
				}
			}
		},
	}
}
