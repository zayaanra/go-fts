package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var ftsCmd = &cobra.Command{
	Use:	"fts",
	Short:	"FTS is a CLI tool for sending and receiving files to other computers.",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func Execute() {
	if err := ftsCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "An error occured while executing FTS: '%s'\n", err)
		os.Exit(1)
	}
}