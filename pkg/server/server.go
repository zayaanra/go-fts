package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/zayaanra/go-fts/pkg/api"
)

type RelayService interface {
	Send() error

	Receive() error

	Close() error
}

type Room struct {
	// List of connections connected to this room
	connections []*net.Conn
}

type Server struct {
	// The address for this server
	ip		 string
	port	 string
	
	// The handler for this server
	listener net.Listener

	// Map each session id to their room
	rooms 	 map[string]*Room

	C		chan *api.Message
	quit	chan bool
}

func NewServer(port string) (RelayService, error) {
	addr := fmt.Sprintf(":" + port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("Failed to create listener on %s: %w", addr, err)
	}

	server := &Server{
		ip: 	  listener.Addr().(*net.TCPAddr).IP.String(),
		port:	  strconv.Itoa(listener.Addr().(*net.TCPAddr).Port),
		listener: listener,
		rooms: 	  map[string]*Room{},
		C:		  make(chan *api.Message),
		quit:	  make(chan bool, 1),
	}

	go func() {
		for {
			select {
			case <-server.quit:
				close(server.quit)
				return
			default:
				conn, err := listener.Accept()
				if err != nil {
					log.Printf("Failed to accept incoming connection: %s", err)
				}

				go handleConnection(conn, server)
			}

		}
	}()

	return server, nil
}

func (s *Server) Send() error {
	return nil
}

func (s *Server) Receive() error {
	return nil
}

func (s *Server) Close() error {
	s.quit <- true

	err := s.listener.Close()
	if err != nil {
		log.Fatalf("Failed to close RelayServer: %s", err)
		return err
	}

	close(s.C)
	return nil
}

func handleConnection(conn net.Conn, s *Server) {
	defer conn.Close()

	var msg api.Message

	decoded := json.NewDecoder(conn)
	if err := decoded.Decode(&msg); err != nil {
		log.Println("Decode error:", err)
		return
	}

	log.Println("Session: ", msg.Session_ID)
	log.Println("Data length: ", len(msg.Data))

}