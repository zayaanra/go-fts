package main

import (
	"fmt"
	"log"
	"os"
	
	"github.com/zayaanra/go-fts/cmd/fts"

	"github.com/joho/godotenv"
)

func main() {
	 err := godotenv.Load()
	 if err != nil {
		log.Fatalf("Error loading .env file")
	 }

	rootCmd := fts.RootCommand(os.Getenv("IP_ADDRESS"))

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}