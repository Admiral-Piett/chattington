package servers

import (
	"github.com/Admiral-Piett/chat-telnet/clients"
	cache2 "github.com/patrickmn/go-cache"
	"log"
	"net"
)

type Server struct {
	Listener net.Listener
}

func NewServer() (Server, error) {
	l, err := net.Listen("tcp", ":2000")
	if err != nil {
		return Server{}, err
	}
	server := Server{
		Listener: l,
	}
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
	c.Set(clients.PRIVATE_ROOMS, map[string][]*clients.Client{}, cache2.NoExpiration)
	return c
}
