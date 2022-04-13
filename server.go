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
    clients  map[string]*Client     // Currently connected clients.
    rooms    map[string][]*Client    // Running chat rooms.
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
        rooms:    map[string][]*Client{},
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

func (s *Server) parseResponse(cmd string, client *Client) string {
    value := ""
    // Route any commands to their relevant methods
    cmdIndex := strings.IndexByte(cmd, ' ')
    if cmdIndex > 0 {
        value = strings.TrimSpace(cmd[cmdIndex:])
        cmd = cmd[:cmdIndex]
    }
    switch {
    case cmd == "\\list":
        return s.listRooms()
    case cmd == "\\name" && value != "":
        return s.changeClientName(value, client)
    case cmd == "\\create" && value != "":
        return s.createRoom(value, client)
    case cmd == "\\join" && value != "":
        log.Printf("Join Command: %s\n", value)
        //TODO - leave current room and join another
        return ""
    }
    // TODO - \\leave - leave room
    // TODO - \\whoami - Show name and current room?
    // TODO - \\create-private - private room??? - How would that even work?
    // TODO - broadcast message to room
    // TODO - Direct message?
    return fmt.Sprintf("Invalid command: `%s`", cmd)
}

func (s *Server) createClient(conn net.Conn) *Client {
    log.Printf("Accepting new connection from address %v, total clients: %v\n", conn.RemoteAddr().String(), len(s.clients)+1)

    s.mutex.Lock()
    defer s.mutex.Unlock()

    // Generate a semi-random id for ourselves, using a `---` pattern to start.  We can lean on this for now, to
    //identify users who haven't yet given their name but still access the clients by key.
    intialName := fmt.Sprintf("%v", time.Now().Unix())

    client := &Client{
        conn:   conn,
        name:   intialName,
        writer: conn,
    }

    s.clients[intialName] = client

    nameInstructions := fmt.Sprintf("\nWe've set your user name with a default - `%s`\nIf you'd like to reset it, please use the '\\name' command.\n\n%s> ", intialName, client.name)

    client.WriteString(nameInstructions)
    return client
}

func (s *Server) changeClientName(name string, client *Client) string {
    log.Printf("User: %s is becoming -> %s\n", client.name, name)
    s.mutex.Lock()
    defer s.mutex.Unlock()

    // Replace original client name and mapping in our client list with the updated name.
    s.clients[name] = client
    delete(s.clients, client.name)

    // We can still change this name value after it's been set back in the clients pool since it's a pointer.
    client.name = name
    return "User name reset"
}

func (s *Server) listRooms() string {
    log.Println("Listing rooms")
    if len(s.rooms) < 1 {
        return "No rooms yet - make one!"
    }
    roomString := ""
    for name, members := range(s.rooms) {
        roomString = roomString + fmt.Sprintf("Room: %s\nMembers:\n", name)
        for _, client := range members {
            roomString = roomString + fmt.Sprintf("\t%s\n", client.name)
        }
    }
    return fmt.Sprintf("Current rooms: \n%s", roomString)
}

func (s *Server) createRoom(roomName string, client *Client) string {
    if s.rooms[roomName] != nil{
        return "Room already exists - use `\\join` to join the chat."
    }
    log.Printf("Creating Room: %s\n", roomName)
    s.rooms[roomName] = []*Client{client}

    // Leave any existing rooms this user is in since you can only be in 1.
    s.leaveRoom(client.CurrentRoom, client)

    client.CurrentRoom = roomName
    return fmt.Sprintf("New room created: %s", roomName)
}

func (s *Server) leaveRoom(roomName string, client *Client) {
    // Return as a no-op here if you didn't give a roomName (you may not have entered a room yet).
    if roomName == "" {
        return
    }

    s.mutex.Lock()
    defer s.mutex.Unlock()

    // Re-create the list of clients in the given room, minus whatever client we have in play since they are leaving.
    prunedList := []*Client{}
    for _, c := range(s.rooms[roomName]) {
       if client != c {
           prunedList = append(prunedList, client)
        }
    }
    s.rooms[roomName] = prunedList

    /// If the room no longer has anyone in it after this user has been removed then delete it.
    if len(s.rooms[roomName]) == 0 {
       delete(s.rooms, roomName)
    }

    // TODO - Consider the map version below relying on this data structure
    //      {<room name>: {<client name>: <client pointer>} }
    //  I'd probably need a client Id or something static to use as a key instead of the client.name to deal with this,
    //  since the names can change and that could leave you in 2 rooms.
    //// Delete this client from the rooms: clients mappings.
    //if s.rooms[roomName][clientName] != nil {
    //    delete(s.rooms[roomName], clientName)
    //}
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
