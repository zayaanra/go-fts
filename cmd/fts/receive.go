package fts

import (
	"github.com/spf13/cobra"
)

func ReceiveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "receive file_path",
		Short: "Receive a file",
		Long:  "Receive a file from a machine",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
}
