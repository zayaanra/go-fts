package cmd

import (
	"fmt"
	"errors"
	"net/http"
	"github.com/spf13/cobra"
)

var receiveCmd = &cobra.Command{
	Use:	"receive",
	Short: 	"Receive a file",
	Long: 	"Receive a file from a computer",
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("Receiving...")
		go func () {
			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				fmt.Printf("server: %s /\n", r.Method)		
			})
			server := http.Server{
				Addr:	fmt.Sprintf(":%d", 8000),
				Handler: mux,
			}
			if err := server.ListenAndServe(); err != nil {
				if !errors.Is(err, http.ErrServerClosed) {
					fmt.Printf("error running http server: %s\n", err)
				}
			}
		}()
	},
}

func init() {
	ftsCmd.AddCommand(receiveCmd)
}