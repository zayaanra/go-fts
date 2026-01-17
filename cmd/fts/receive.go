package fts

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/zayaanra/go-fts/peer"
)

func ReceiveCommand(ip string) *cobra.Command {
	return &cobra.Command{
		Use:   "receive [output-path]",
		Short: "Receive a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// outputPath := args[0]

			fmt.Println("Enter the code shared with you:")
		
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()

			passphrase := scanner.Text()
			sessionID := strings.Split(passphrase, "-")[0]

			p := peer.NewPeer(peer.PAKE_RESPONDER, sessionID, passphrase)
			if err := p.Rendevous(ip); err != nil {
				return err
			}
			defer p.Close()

			if err := p.Listen(); err != nil {
				return fmt.Errorf("Either PAKE or something else failed: %w", err)
			}

			color.Cyan("\nFile received successfully!")
			return nil
		},
	}
}
