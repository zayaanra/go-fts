package fts

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/zayaanra/go-fts/pkg/peer"
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

			if err := p.ListenWS(); err != nil {
				return fmt.Errorf("Either PAKE or something else failed: %w", err) 
			}

			// 1. Create the destination file
			// f, err := os.Create(outputPath)
			// if err != nil {
			// 	return fmt.Errorf("failed to create file: %w", err)
			// }
			// defer f.Close()

			// 2. Setup Progress Bar (Note: You might need to send the 
			// file size over the network first to make this bar accurate)
			// bar := progressbar.Default(-1, "Downloading") 

			// 3. Wrap file with progress bar
			// proxyWriter := io.MultiWriter(f, bar)

			// fmt.Printf("Receiving data into %s...\n", outputPath)
			
			// if err := p.ReceiveData(proxyWriter); err != nil {
			// 	return err
			//}

			color.Green("\nFile received successfully!")
			return nil
		},
	}
}
