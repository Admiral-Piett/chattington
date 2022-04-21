package servers

import (
	"chat-telnet/clients"
	"fmt"
	cache2 "github.com/patrickmn/go-cache"
	"log"
	"net"
	"os"
)

type Server struct {
	Listener net.Listener
}

func NewServer() (Server, error) {
	address := fmt.Sprintf("%s:%s", os.Getenv("IP_ADDRESS"), os.Getenv("PORT"))
	l, err := net.Listen("tcp", address)
	if err != nil {
		return Server{}, err
	}
	server := Server{
		Listener: l,
	}
	log.Printf("Starting chat-telnet server on address: %s", address)
	return server, nil
}

func (s *Server) Close() {
	s.Listener.Close()
}

func (s *Server) Start() error {
	cache := NewChatCache() // pointer to our global cache
	for {
		// Wait for a connection.
		conn, err := s.Listener.Accept()
		if err != nil {
			return err
		}

		// If we fail to generate a client when the user connects log and close the connection, letting them try again.
		//	Keep the server going though to continue listening.
		err = clients.GenerateNewClient(conn, cache)
		if err != nil {
			log.Println(err)
			conn.Close()
		}
	}
}

// This should create a thread-safe cache so we should be about to pound it with go routines all we want.
func NewChatCache() *cache2.Cache {
	c := cache2.New(cache2.NoExpiration, cache2.NoExpiration)
	c.Set(clients.CLIENTS, map[string]*clients.Client{}, cache2.NoExpiration)
	c.Set(clients.ROOMS, map[string][]*clients.Client{}, cache2.NoExpiration)
	return c
}
