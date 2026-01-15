package fts

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/sethvargo/go-diceware/diceware"
	"github.com/spf13/cobra"

	"github.com/zayaanra/go-fts/pkg/api"
	"github.com/zayaanra/go-fts/pkg/crypt"
	"github.com/zayaanra/go-fts/pkg/peer"
)

func SendCommand(ip string) *cobra.Command {
	return &cobra.Command{
		Use:   "send [file-path]",
		Short: "Send a file",
		Long:  "Send a file to a machine",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			fileInfo, err := os.Stat(filePath)
			if err != nil {
				return fmt.Errorf("Could not find file: %w", err)
			}

			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			list, _ := diceware.Generate(5)
			sessionID := uuid.New().String()[:5]
			passphrase := sessionID + "-" + strings.Join(list, "-")

			p := peer.NewPeer(peer.PAKE_INITIATOR, sessionID, passphrase)
			if err := p.Rendevous(ip); err != nil {
				return fmt.Errorf("Rendezvous failed: %w", err)
			}
			defer p.Close()

			green := color.New(color.FgGreen).Add(color.Bold)
			fmt.Println("On the receiving machine, run the receive command and enter the following code:")
			green.Printf("> %s\n\n", passphrase)

			// bar := progressbar.DefaultBytes(
			// 	fileInfo.Size(),
			// 	"Sending file",
			// )

			if err := p.ListenWS(); err != nil {
				return fmt.Errorf("Either PAKE or something else failed: %w", err)
			}

			conn, err := net.Dial("tcp", p.ReceiverIP)
			if err != nil {
				return fmt.Errorf("Failed to connect to receiver: %w", err)
			}
			defer conn.Close()


			fileSize := make([]byte, 8)
			binary.BigEndian.PutUint64(fileSize, uint64(fileInfo.Size()))

			fileData := make([]byte, fileInfo.Size())
			file.Read(fileData)

			fileMsg, _ := json.Marshal(
				api.File{
					Length: fileSize, 
					Data: fileData,
				},
			)
			encrypted, _ := crypt.EncryptAES(fileMsg, p.Session.Key)
			conn.Write(encrypted)

			// proxyReader := io.TeeReader(file, bar)
			// if err := p.StreamData(proxyReader); err != nil {
			// 	return fmt.Errorf("Transfer failed: %w", err)
			// }

			color.Cyan("\nTransfer complete!")
			return nil

			// fileInfo, _ := os.Stat(args[0])
			// totalSize := fileInfo.Size()

			// bar := progressbar.DefaultBytes(
			// 	totalSize,
			// 	"sending",
			// )

			// file, _ := os.Open(args[0])
			// defer file.Close()

			// fmt.Printf("Sending '%s' -> '%s", args[0], p.ReceiverIP)
		},
	}
}
