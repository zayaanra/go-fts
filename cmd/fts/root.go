package fts

import (
	"fmt"

	"github.com/spf13/cobra"
)

func RootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "fts",
		Short: "FTS is a CLI tool for sending and receiving files to other machines.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "Welcome to Go-FTS. Please provide --help for more information regarding this tool.")
		},
	}

	cmd.AddCommand(SendCommand())
	cmd.AddCommand(ReceiveCommand())

	return cmd
}
