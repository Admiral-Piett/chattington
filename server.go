package main

import (
    "fmt"
    "io"
    "log"
    "net"
    "strings"
    "sync"
    "time"
)

type Server struct {
    listener net.Listener
    clients map[string]*client
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
        clients: map[string]*client{},
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

    // Generate a semi-random id for ourselves, using a `---` pattern to start.  We can lean on this for now, to
    //identify users who haven't get given their name but still access the clients by key.
    intialName := fmt.Sprintf("---%v", time.Now().Unix())

    client := &client{
        conn: conn,
        writer: NewCommandWriter(conn),
        name: intialName,
    }

    s.clients[intialName] = client

    client.writer.WriteString("Please input your user name: ")
    return client
}

func (s *Server) removeConnection(client *client) {
    delete(s.clients, client.name)
    log.Printf("Removed connection %v from pool, total clients: %v\n", client.conn.RemoteAddr().String(), len(s.clients))
}

func (s *Server) listen(client *client) {
    r := NewCommandReader(client.conn)
    defer s.removeConnection(client)

    for {
       input, err := r.Read()
       if err != nil && err != io.EOF {
           log.Printf("Read error: %v\n", err)
           client.writer.WriteString(fmt.Sprintf("%s\n", err))
       }

       if input != "" {
           if strings.HasPrefix(client.name, "---") {
               // Replace original client name and mapping in our client list with the updated value.
               s.clients[input] = client
               delete(s.clients, client.name)

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
