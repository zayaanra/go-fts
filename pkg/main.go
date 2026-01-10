package main

import (
	"log"
	"net/http"

	"github.com/zayaanra/go-fts/pkg/server"
)

func main() {
	hub := server.NewMailbox()
	go hub.Run()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		server.ServeWS(hub, w, r)
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}