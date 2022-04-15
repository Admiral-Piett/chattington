package servers

import (
    "github.com/Admiral-Piett/chat-telnet/clients"
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
    for {
        // Wait for a connection.
        conn, err := s.Listener.Accept()
        if err != nil {
            return err
        }

        clients.GenerateNewClient(conn)
    }
}
