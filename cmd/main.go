package main

import (
	"fmt"
	"os"

	"github.com/zayaanra/go-fts/cmd/fts"
)

func main() {
	rootCmd := fts.RootCommand()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}