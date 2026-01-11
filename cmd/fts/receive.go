package fts

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zayaanra/go-fts/pkg/peer"
)

func ReceiveCommand(ip string) *cobra.Command {
	return &cobra.Command{
		Use:   "receive [file-path]",
		Short: "Receive a file",
		Long:  "Receive a file from a machine",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Enter the code shared with you:")

			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()

			passphrase := scanner.Text()
			sessionID := strings.Split(passphrase, "-")[0]
			
			p := peer.NewPeer(peer.PAKE_RESPONDER, sessionID, passphrase)
			if err := p.Rendevous(ip); err != nil {
				log.Fatal(err)
			}
			defer p.Close()
			
			if err := p.ListenWS(); err != nil {
				log.Fatal(err)
			}
		},
	}
}
