package fts

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/sethvargo/go-diceware/diceware"
	"github.com/spf13/cobra"

	"github.com/zayaanra/go-fts/pkg/peer"
)

func SendCommand(ip string) *cobra.Command {
	return &cobra.Command{
		Use:   "send [file-path]",
		Short: "Send a file",
		Long:  "Send a file to a machine",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: Return errors up the chain and handle them in the RunE field of the Cobra command

			list, _ := diceware.Generate(5)
			sessionID := uuid.New().String()[:5]
			passphrase := sessionID + "-" + strings.Join(list, "-")

			p := peer.NewPeer(peer.PAKE_INITIATOR, sessionID, passphrase)
			if err := p.Rendevous(ip); err != nil {
				log.Fatal(err)
			}
			defer p.Close()

			fmt.Println("On the receiving machine, run the receive command and enter the following code:")
			fmt.Println(passphrase + "\n")

			// TODO: What if file is too large to be stored in memory? (can stream data to receiver)
			data, err := os.ReadFile(args[0])
			if err != nil {
						log.Fatal(err)
			}
			p.FileData = data

			if err := p.ListenWS(); err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Sending '%s' -> '%s", args[0], p.ReceiverIP)
		},
	}
}
