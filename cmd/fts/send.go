package fts

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/schollz/progressbar/v3"
	"github.com/sethvargo/go-diceware/diceware"
	"github.com/spf13/cobra"

	"github.com/zayaanra/go-fts/crypt"
	"github.com/zayaanra/go-fts/peer"
)

func SendCommand(ip string) *cobra.Command {
	return &cobra.Command{
		Use:   "send [file-path]",
		Short: "Send a file",
		Long:  "Send a file to a machine",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			fileMetadata, _ := os.Stat(filePath)

			list, _ := diceware.Generate(5)
			sessionID := uuid.New().String()[:5]
			passphrase := sessionID + "-" + strings.Join(list, "-")

			p := peer.NewPeer(peer.PAKE_INITIATOR, sessionID, passphrase)
			if err := p.Rendevous(ip); err != nil {
				return fmt.Errorf("Rendezvous failed: %w", err)
			}
			defer p.Close()

			fmt.Println("On the receiving machine, run the receive command and enter the following code:")
			green := color.New(color.FgGreen).Add(color.Bold)
			green.Printf("> %s\n\n", passphrase)

			bar := progressbar.NewOptions64(
				fileMetadata.Size(),
				progressbar.OptionSetDescription("Sending file"),
				progressbar.OptionShowBytes(true),
				progressbar.OptionSetWidth(40),
				progressbar.OptionThrottle(65*time.Millisecond),
				progressbar.OptionClearOnFinish(),
			)

			if err = p.Listen(); err != nil {
				return fmt.Errorf("Either PAKE or something else failed: %w", err)
			}

			conn, err := net.Dial("tcp", p.ReceiverIP)
			if err != nil {
				return fmt.Errorf("Failed to connect to receiver: %w", err)
			}
			defer conn.Close()
			
			buf := make([]byte, 32*1024)
			for {
				n, err := file.Read(buf)
				if n > 0 {
					encrypted, err := crypt.EncryptAES(buf[:n], p.Session.Key)
					if err != nil {
						return err
					}

					var lenBuf [8]byte
					binary.BigEndian.PutUint64(lenBuf[:], uint64(len(encrypted)))
					conn.Write(lenBuf[:])

					conn.Write(encrypted)

					bar.Add(n)
				}

				if err == io.EOF {
					break
				}
				if err != nil {
					return err
				}
			}
			bar.Finish()

			color.Cyan("\nTransfer complete!")

			return nil
		},
	}
}
