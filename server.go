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
    clients  map[string]*Client     // Currently connected clients.  {<client.name> : *Client}
    rooms    map[string][]*Client    // Running chat rooms.  {<room_name> : [*Client, ...]}
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

// Here we will handle any commands and return anything we want to send back to the client.  If we want this to
//  be "broadcast" (to every client in the chat room) we can answer `true` with the accompanying bool.  Most methods
// should actually return their own values to decide whether their message gets broadcast or not, so we can send
//error messages privately, etc.
func (s *Server) parseResponse(cmd string, client *Client) (string, bool) {
    value := ""
    // Route any commands to their relevant methods
    cmdIndex := strings.IndexByte(cmd, ' ')
    if cmdIndex > 0 {
        value = strings.TrimSpace(cmd[cmdIndex:])
        cmd = cmd[:cmdIndex]
    }
    switch {
    case cmd == "\\name" && value != "":
        return s.changeClientName(value, client)
    case cmd == "\\create" && value != "":
        return s.createRoom(value, client)
    case cmd == "\\join" && value != "":
        return s.joinRoom(value, client)
    case cmd == "\\list" && value != "":
        return s.listRooms()
    case cmd == "\\leave":
        s.leaveRoom(client.CurrentRoom, client)
        s := fmt.Sprintf("Left room %s", client.CurrentRoom)
        client.CurrentRoom = ""
        // TODO - HERE - this isn't working since we've already left the room before doing it.
        //  I don't really want to but we may have to fire a broadcast from in here.
        return s, true
    case cmd == "\\list":  // list the members of your current room
        return s.listMembers(client.CurrentRoom)
    case cmd == "\\list-rooms":
        return s.listRooms()
    case cmd == "\\whoami":
        return s.displayClientStats(client)
    }
    // TODO - \\create-private - private room??? - How would that even work?
    // TODO - Direct message?
    return fmt.Sprintf("Invalid command: `%s`", cmd), false
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

    nameInstructions := fmt.Sprintf("\nWe've set your user name with a default - `%s`\nIf you'd like to reset it, please use the '\\name' command.\n\n", intialName)

    client.WriteString(nameInstructions)
    return client
}

func (s *Server) changeClientName(name string, client *Client) (string, bool) {
    msg := fmt.Sprintf("User: %s has become -> %s", client.name, name)
    s.mutex.Lock()
    defer s.mutex.Unlock()

    // Replace original client name and mapping in our client list with the updated name.
    s.clients[name] = client
    delete(s.clients, client.name)

    // We can still change this name value after it's been set back in the clients pool since it's a pointer.
    client.name = name
    return msg, true
}

func (s *Server) displayClientStats(client *Client) (string, bool) {
    currentRoom := client.CurrentRoom
    if currentRoom == "" {
        currentRoom = "None"
    }
    return fmt.Sprintf("\nClient Name: %s\nCurrent Room: %s", client.name, currentRoom), false
}

func (s *Server) listRooms() (string, bool) {
    if len(s.rooms) < 1 {
        return "No rooms yet - make one!", false
    }
    roomString := ""
    for name, members := range(s.rooms) {
        roomString = roomString + fmt.Sprintf("  Room: %s\n  Members:\n", name)
        for _, client := range members {
            roomString = roomString + fmt.Sprintf("\t%s\n", client.name)
        }
    }
    return fmt.Sprintf("\nCurrent rooms: \n%s", roomString), false
}

func (s *Server) listMembers(roomName string) (string, bool) {
    if s.rooms[roomName] == nil {
        return fmt.Sprintf("No such room %s!", roomName), false
    }
    roomString := ""
    for _, client := range(s.rooms[roomName]) {
        roomString = roomString + fmt.Sprintf("\t%s\n", client.name)
    }
    return fmt.Sprintf("\nCurrent Members:\n%s", roomString), false
}

func (s *Server) createRoom(roomName string, client *Client) (string, bool) {
    if s.rooms[roomName] != nil{
        return "Room already exists - use `\\join` to join the chat.", false
    }
    s.mutex.Lock()
    defer s.mutex.Unlock()

    log.Printf("Creating Room: %s\n", roomName)
    s.rooms[roomName] = []*Client{client}

    // Leave any existing rooms this user is in since you can only be in 1.
    s.leaveRoom(client.CurrentRoom, client)

    client.CurrentRoom = roomName
    return fmt.Sprintf("New room created: %s", roomName), false
}

func (s *Server) joinRoom(roomName string, client *Client) (string, bool) {
    if s.rooms[roomName] == nil {
        return fmt.Sprintf("Room `%s` doesn't exist - try creating it with `\\create`", roomName), false
    }
    if client.CurrentRoom == roomName {
        return fmt.Sprintf("You're already in room %s!", roomName), false
    }
    s.mutex.Lock()
    defer s.mutex.Unlock()

    s.rooms[roomName] = append(s.rooms[roomName], client)

    // Leave any existing rooms this user is in since you can only be in 1.
    s.leaveRoom(client.CurrentRoom, client)

    client.CurrentRoom = roomName
    return fmt.Sprintf("%s has entered: %s", client.name, roomName), true
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
           prunedList = append(prunedList, c)
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
    s.mutex.Lock()
    defer s.mutex.Unlock()

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
            client.WriteResponse(input, nil)
        }

        // NOTE: This fires only when the Client kills its connection
        if err == io.EOF {
            break
        }

        if input != "" {
            if strings.HasPrefix(input, "\\") {
                response, toBroadcast := s.parseResponse(input, client)
                if response != "" {
                    if toBroadcast {
                        go s.broadcastToRoom(response, client)
                    } else {
                        client.WriteResponse(response, nil)
                    }
                }
            } else if client.CurrentRoom != "" {
                go s.broadcastToRoom(input, client)
            } else {
                client.WriteResponse(input, nil)
            }
        }

    }
}

func (s *Server) broadcastToRoom(message string, client *Client) {
    // If no one is in the room I'm in then just send it to myself.
    if len(s.rooms[client.CurrentRoom]) < 1 {
        client.WriteResponse(message, nil)
    }
    for _, c := range s.rooms[client.CurrentRoom] {
        c.WriteResponse(message, client.name)
    }
}
