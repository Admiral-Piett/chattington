package main

import (
    "bufio"
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
    clients  map[string]*Client
    rooms    map[string][]string
    mutex    *sync.Mutex
}

func NewServer() (Server, error) {
    l, err := net.Listen("tcp", ":2000")
    if err != nil {
        log.Fatal(err)
    }
    server := Server{
        listener: l,
        clients:  map[string]*Client{},
        rooms:    map[string][]string{},
        mutex:    &sync.Mutex{},
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

func (s *Server) parseResponse(entry string, client *Client) string {
    // Route any commands to their relevant methods
    cmdIndex := strings.IndexByte(entry, ' ')
    if cmdIndex < 0 {
        return fmt.Sprintf("Invalid command: `%s`", entry)
    }
    cmd := entry[:cmdIndex]
    value := strings.TrimSpace(entry[cmdIndex:])
    switch {
    case cmd == "\\name":
        log.Printf("Name Command: %s\n", entry)
        s.changeClientName(value, client)
        return "User name reset"
    case cmd == "\\list-rooms":
        log.Printf("List Rooms Command: %s\n", entry)
        return ""
    case cmd == "\\create":
        log.Printf("Create Command: %s\n", entry)
        return ""
    case cmd == "\\join":
        log.Printf("Join Command: %s\n", entry)
        return ""
    }
    return ""
}

func (s *Server) changeClientName(name string, client *Client){
    s.mutex.Lock()
    defer s.mutex.Unlock()

   // Replace original client name and mapping in our client list with the updated name.
   s.clients[name] = client
   delete(s.clients, client.name)

   // We can still change this name value after it's been set back in the clients pool since it's a pointer.
   client.name = name
}

func (s *Server) createClient(conn net.Conn) *Client {
    log.Printf("Accepting new connection from address %v, total clients: %v\n", conn.RemoteAddr().String(), len(s.clients)+1)

    s.mutex.Lock()
    defer s.mutex.Unlock()

    // Generate a semi-random id for ourselves, using a `---` pattern to start.  We can lean on this for now, to
    //identify users who haven't get given their name but still access the clients by key.
    intialName := fmt.Sprintf("%v", time.Now().Unix())

    client := &Client{
        conn:   conn,
        name:   intialName,
        writer: conn,
    }

    s.clients[intialName] = client

    nameInstructions := fmt.Sprintf("\nWe've set your user name with a default - `%s`\nIf you'd like to reset it, please use the '\\name' command.\n\n", intialName)

    client.WriteString(nameInstructions)
    return client
}

func (s *Server) removeConnection(client *Client) {
    delete(s.clients, client.name)
    log.Printf("Removed connection %v from pool, total clients: %v\n", client.conn.RemoteAddr().String(), len(s.clients))
}

func (s *Server) listen(client *Client) {
    r := bufio.NewReader(client.conn)
    defer s.removeConnection(client)

    for {
        input, err := Read(r)
        if err != nil && err != io.EOF {
            log.Printf("Read error: %v\n", err)
            client.WriteResponse(input)
        }

        // NOTE: This fires only when the Client kills its connection
        if err == io.EOF {
            break
        }

        if input != "" {
            if strings.HasPrefix(input, "\\") {
                response := s.parseResponse(input, client)
                if response != "" {
                    client.WriteResponse(response)
                }
            } else {
                client.WriteResponse(input)
            }
        }

    }
}
