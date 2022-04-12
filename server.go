package main

import (
    "fmt"
    "io"
    "log"
    "net"
    "sync"
)

type Server struct {
    listener net.Listener
    clients []*client
    mutex   *sync.Mutex
}

type client struct {
    conn   net.Conn
    name   string
    writer *ChatWriter
}

func NewServer() (Server, error){
    l, err := net.Listen("tcp", ":2000")
    if err != nil {
        log.Fatal(err)
    }
    server := Server{
        listener: l,
        mutex: &sync.Mutex{},
    }
    return server, nil
}

func (s *Server) Close() {
    s.listener.Close()
}


//TODO - handle all fatal errors
func (s *Server) Start() {
    for {
        // Wait for a connection.
        conn, err := s.listener.Accept()
        if err != nil {
            log.Fatal(err)
        }

        client := s.createClient(conn)
        go s.listen(client)
    }
}

func (s *Server) createClient(conn net.Conn) *client {
    log.Printf("Accepting new connection from address %v, total clients: %v\n", conn.RemoteAddr().String(), len(s.clients)+1)

    s.mutex.Lock()
    defer s.mutex.Unlock()

    client := &client{
        conn: conn,
        writer: NewCommandWriter(conn),
    }

    // TODO - append to map instead
    s.clients = append(s.clients, client)

    client.writer.WriteString("Please input your user name: ")
    return client
}

func (s *Server) removeConnection(client *client) {
    log.Printf("Removing connection %v from pool, total clients: %v\n", client.conn.RemoteAddr().String(), len(s.clients)-1)
}

func (s *Server) listen(client *client) {
    r := NewCommandReader(client.conn)
    defer s.removeConnection(client)

    for {
       input, err := r.Read()
       if err != nil && err != io.EOF {
           log.Printf("Read error: %v\n", err)
       }

       if input != "" {
           if client.name == "" {
               client.name = input
               client.writer.WriteString(fmt.Sprintf("User name set: %s\n", input))
           } else {
               client.writer.WriteString(fmt.Sprintf("%s\n", input))
           }
       }

       // NOTE: This fires only when the client kills its connection
       if err == io.EOF {
           break
       }
    }
}
