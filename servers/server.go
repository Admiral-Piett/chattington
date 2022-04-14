package servers

import (
    "bufio"
    "fmt"
    "github.com/Admiral-Piett/chat-telnet/clients"
    "io"
    "log"
    "net"
    "strings"
    "sync"
)

type Server struct {
    listener net.Listener
    clients  map[string]*clients.Client   // Currently connected clients.  {<client.name> : *Client}
    rooms    map[string][]*clients.Client // Running chat rooms.  {<room_name> : [*Client, ...]}
    mutex    *sync.Mutex
}

func NewServer() (Server, error) {
    l, err := net.Listen("tcp", ":2000")
    if err != nil {
        log.Fatal(err)
    }
    server := Server{
        listener: l,
        clients:  map[string]*clients.Client{},
        rooms:    map[string][]*clients.Client{},
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

        c := s.createClient(conn)
        go s.listen(c)
    }
}

// Here we will handle any commands and return anything we want to send back to the client.  If we want this to
//  be "broadcast" (to every client in the chat room) we can answer `true` with the accompanying bool.  Most methods
// should actually return their own values to decide whether their message gets broadcast or not, so we can send
//error messages privately, etc.
func (s *Server) parseResponse(cmd string, c *clients.Client) (string, bool) {
    value := ""
    // Route any commands to their relevant methods
    cmdIndex := strings.IndexByte(cmd, ' ')
    if cmdIndex > 0 {
        value = strings.TrimSpace(cmd[cmdIndex:])
        cmd = cmd[:cmdIndex]
    }
    switch {
    case cmd == "\\name" && value != "":
        return s.changeClientName(value, c)
    case cmd == "\\create" && value != "":
        return s.createRoom(value, c)
    case cmd == "\\join" && value != "":
        return s.joinRoom(value, c)
    case cmd == "\\list" && value != "":
        return s.listRooms()
    case cmd == "\\leave":
        roomName := c.CurrentRoom
        s.leaveRoom(c.CurrentRoom, c)

        c.CurrentRoom = ""
        return fmt.Sprintf("You have left room %s", roomName), false
    case cmd == "\\list":  // list the members of your current room
        return s.listMembers(c.CurrentRoom)
    case cmd == "\\list-rooms":
        return s.listRooms()
    case cmd == "\\whoami":
        return s.displayClientStats(c)
    }
    // TODO - \\create-private - private room??? - How would that even work?
    // TODO - Direct message?
    return fmt.Sprintf("Invalid command: `%s`", cmd), false
}

func (s *Server) createClient(conn net.Conn) *clients.Client {
    log.Printf("Accepting new connection from address %v, total clients: %v\n", conn.RemoteAddr().String(), len(s.clients)+1)

    s.mutex.Lock()
    defer s.mutex.Unlock()

    c, name := clients.NewClient(conn)
    s.clients[name] = c

    nameInstructions := fmt.Sprintf("\nWe've set your user name with a default - `%s`\nIf you'd like to reset it, please use the '\\name' command.\n\n", name)

    c.WriteString(nameInstructions)
    return c
}

func (s *Server) changeClientName(name string, c *clients.Client) (string, bool) {
    msg := fmt.Sprintf("User: %s has become -> %s", c.Name, name)
    s.mutex.Lock()
    defer s.mutex.Unlock()

    // Replace original client name and mapping in our client list with the updated name.
    s.clients[name] = c
    delete(s.clients, c.Name)

    // We can still change this name value after it's been set back in the clients pool since it's a pointer.
    c.Name = name
    return msg, true
}

func (s *Server) displayClientStats(c *clients.Client) (string, bool) {
    currentRoom := c.CurrentRoom
    if currentRoom == "" {
        currentRoom = "None"
    }
    return fmt.Sprintf("\nClient Name: %s\nCurrent Room: %s", c.Name, currentRoom), false
}

func (s *Server) listRooms() (string, bool) {
    if len(s.rooms) < 1 {
        return "No rooms yet - make one!", false
    }
    roomString := ""
    for name, members := range(s.rooms) {
        roomString = roomString + fmt.Sprintf("  Room: %s\n  Members:\n", name)
        for _, c := range members {
            roomString = roomString + fmt.Sprintf("\t%s\n", c.Name)
        }
    }
    return fmt.Sprintf("\nCurrent rooms: \n%s", roomString), false
}

func (s *Server) listMembers(roomName string) (string, bool) {
    if s.rooms[roomName] == nil {
        return fmt.Sprintf("No such room %s!", roomName), false
    }
    roomString := ""
    for _, c := range(s.rooms[roomName]) {
        roomString = roomString + fmt.Sprintf("\t%s\n", c.Name)
    }
    return fmt.Sprintf("\nCurrent Members:\n%s", roomString), false
}

func (s *Server) createRoom(roomName string, c *clients.Client) (string, bool) {
    if s.rooms[roomName] != nil{
        return "Room already exists - use `\\join` to join the chat.", false
    }
    // Defer first since this since it also locks the mutex and this should run last.
    // Leave any existing rooms this user is in since you can only be in 1.
    defer s.leaveRoom(c.CurrentRoom, c)

    s.mutex.Lock()
    defer s.mutex.Unlock()

    log.Printf("Creating Room: %s\n", roomName)
    s.rooms[roomName] = []*clients.Client{c}

    c.CurrentRoom = roomName
    return fmt.Sprintf("New room created: %s", roomName), false
}

func (s *Server) joinRoom(roomName string, c *clients.Client) (string, bool) {
    if s.rooms[roomName] == nil {
        return fmt.Sprintf("Room `%s` doesn't exist - try creating it with `\\create`", roomName), false
    }
    if c.CurrentRoom == roomName {
        return fmt.Sprintf("You're already in room %s!", roomName), false
    }
    // Defer first since this since it also locks the mutex and this should run last.
    // Leave any existing rooms this user is in since you can only be in 1.
    defer s.leaveRoom(c.CurrentRoom, c)

    s.mutex.Lock()
    defer s.mutex.Unlock()

    s.rooms[roomName] = append(s.rooms[roomName], c)

    c.CurrentRoom = roomName
    return fmt.Sprintf("%s has entered: %s", c.Name, roomName), true
}

func (s *Server) leaveRoom(roomName string, c *clients.Client) {
    // Return as a no-op here if you didn't give a roomName (you may not have entered a room yet).
    if roomName == "" {
        return
    }

    s.mutex.Lock()
    defer s.mutex.Unlock()

    // Re-create the list of clients in the given room, minus whatever client we have in play since they are leaving.
    prunedList := []*clients.Client{}
    for _, c := range(s.rooms[roomName]) {
       if c != c {
           prunedList = append(prunedList, c)
        }
    }
    s.rooms[roomName] = prunedList

    /// If the room no longer has anyone in it after this user has been removed then delete it.
    if len(s.rooms[roomName]) == 0 {
       delete(s.rooms, roomName)
    }

    // I hate to do this in here, but I don't really want to pass roomName up through all these methods and
    //their associated conditions when 90% of the time it's going to be what's already on the client.  So
    //leaving this for now.
    go s.broadcastToRoom(fmt.Sprintf("%s has left the room %s.", c.Name, roomName), roomName, c)

    // TODO - Consider the map version below relying on this data structure
    //      {<room name>: {<client name>: <client pointer>} }
    //  I'd probably need a client Id or something static to use as a key instead of the client.name to deal with this,
    //  since the names can change and that could leave you in 2 rooms.
    //// Delete this client from the rooms: clients mappings.
    //if s.rooms[roomName][clientName] != nil {
    //    delete(s.rooms[roomName], clientName)
    //}
}

func (s *Server) removeConnection(c *clients.Client) {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    delete(s.clients, c.Name)
    log.Printf("Removed connection %v from pool, total clients: %v\n", c.Conn.RemoteAddr().String(), len(s.clients))
}

func (s *Server) listen(c *clients.Client) {
    r := bufio.NewReader(c.Conn)
    defer s.removeConnection(c)

    for {
        input, err := clients.Read(r)
        if err != nil && err != io.EOF {
            log.Printf("Read error: %v\n", err)
            c.WriteResponse(input, nil)
        }

        // NOTE: This fires only when the Client kills its connection
        if err == io.EOF {
            break
        }

        if input != "" {
            if strings.HasPrefix(input, "\\") {
                response, toBroadcast := s.parseResponse(input, c)
                if response != "" {
                    if toBroadcast {
                        go s.broadcastToRoom(response, c.CurrentRoom, c)
                    } else {
                        c.WriteResponse(response, nil)
                    }
                }
            } else if c.CurrentRoom != "" {
                go s.broadcastToRoom(input, c.CurrentRoom, c)
            } else {
                c.WriteResponse(input, nil)
            }
        }

    }
}

func (s *Server) broadcastToRoom(message, roomName string, c *clients.Client) {
    // If no one is in the room I'm in then just send it to myself.
    if len(s.rooms[roomName]) < 1 {
        c.WriteResponse(message, nil)
    }
    for _, targetClient := range s.rooms[roomName] {
        targetClient.WriteResponse(message, c.Name)
    }
}
