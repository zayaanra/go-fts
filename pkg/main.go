package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/zayaanra/go-fts/pkg/server"
)

func main() {
	server, err := server.NewServer("8090")
	if err != nil {
		log.Fatalf("Something went wrong: %v", err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig
	log.Println("Shutting down RelayServer...")
	server.Close()
}