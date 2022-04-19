package clients

import (
	"bufio"
	"fmt"
	"github.com/Admiral-Piett/chat-telnet/interfaces"
	"io"
	"log"
	"strings"
	"sync"
	"time"
)

// TODO - explore a cache like redis or BadgerDB?
type ChatMeta struct {
	Clients map[string]*Client   // Currently connected clients.  {<client.name> : *Client}
	Rooms   map[string][]*Client // Running chat rooms.  {<room_name> : [*Client, ...]}
	Mutex   *sync.Mutex
}

var ChatCache = &ChatMeta{
	Clients: map[string]*Client{},
	Rooms:   map[string][]*Client{},
	Mutex:   &sync.Mutex{},
}

type Client struct {
	Writer      interfaces.AbstractIoWriter
	Conn        interfaces.AbstractNetConn
	Name        string
	CurrentRoom string
	Id          string
}

func GenerateNewClient(conn interfaces.AbstractNetConn) {
	log.Printf("Accepting new connection from address %v, total clients: %v\n", conn.RemoteAddr().String(), len(ChatCache.Clients)+1)

	ChatCache.Mutex.Lock()
	defer ChatCache.Mutex.Unlock()

	// Generate a semi-random id and initial name for ourselves, using a time stamp.  For now, this is "unique"
	//  enough since our users are not logging on so quickly that they should collide.
	id := fmt.Sprintf("%v", time.Now().Unix())
	client := &Client{
		Conn:        conn,
		Writer:      conn,
		Name:        id,
		CurrentRoom: "",
		Id:          id,
	}

	ChatCache.Clients[id] = client
	nameInstructions := fmt.Sprintf("\nWe've set your user name with a default - `%s`\nIf you'd like to reset it, please use the '\\name' comman\n\n", id)

	client.WriteString(nameInstructions)
	go client.listen()
}

func (c *Client) WriteString(msg string) error {
	_, err := c.Writer.Write([]byte(msg))

	return err
}

func (c *Client) WriteResponse(msg string, sendingClient interface{}) error {
	// Using `sendingClient` you can send in the string name who the sender of the message is, if it's `nil` we'll
	//  or the same as the target client (`c`) then we can format the response as i.
	prefix := ""
	if sendingClient == nil || sendingClient == c.Name {
		prefix = fmt.Sprintf("%s>", c.Name)
	} else {
		prefix = fmt.Sprintf("%s:", sendingClient)
	}
	// Add chat room response formatting
	msg = fmt.Sprintf("%s %s\n", prefix, msg)
	return c.WriteString(msg)
}

// Should I attach this to a struct?
func Read(r interfaces.AbstractBufioReader) (string, error) {
	value, err := r.ReadString('\n')
	// Looks like can usually expect an io.EOF on connection death here, or potential bad characters, etc.  Either
	//way, log it back to the user, so they can see it and stop putting garbage back (assuming they haven't
	//killed their connection in which case it won't matter).
	if err != nil {
		return "", err
	}

	value = strings.TrimSpace(value)
	return value, err
}

func (c *Client) removeConnection() {
	ChatCache.Mutex.Lock()
	defer ChatCache.Mutex.Unlock()

	delete(ChatCache.Clients, c.Id)
	log.Printf("Removed connection %v from pool, total clients: %v\n", c.Conn.RemoteAddr().String(), len(ChatCache.Clients))
}

func (c *Client) changeClientName(name string) (string, bool) {
	msg := fmt.Sprintf("User: %s has become -> %s", c.Name, name)
	ChatCache.Mutex.Lock()
	defer ChatCache.Mutex.Unlock()

	c.Name = name
	return msg, true
}

func (c *Client) displayClientStats() (string, bool) {
	currentRoom := c.CurrentRoom
	if currentRoom == "" {
		currentRoom = "None"
	}
	return fmt.Sprintf("\nClient Name: %s\nCurrent Room: %s", c.Name, currentRoom), false
}

func (c *Client) listRooms() (string, bool) {
	if len(ChatCache.Rooms) < 1 {
		return "No rooms yet - make one!", false
	}
	roomString := ""
	for name, members := range ChatCache.Rooms {
		roomString = roomString + fmt.Sprintf("  Room: %s\n  Members:\n", name)
		for _, c := range members {
			roomString = roomString + fmt.Sprintf("\t%s\n", c.Name)
		}
	}
	return fmt.Sprintf("\nCurrent rooms: \n%s", roomString), false
}

func (c *Client) listMembers(roomName string) (string, bool) {
	if ChatCache.Rooms[roomName] == nil {
		return fmt.Sprintf("No such room %s!", roomName), false
	}
	roomString := ""
	for _, c := range ChatCache.Rooms[roomName] {
		roomString = roomString + fmt.Sprintf("\t%s\n", c.Name)
	}
	return fmt.Sprintf("\nCurrent Members:\n%s", roomString), false
}

func (c *Client) createRoom(roomName string) (string, bool) {
	if ChatCache.Rooms[roomName] != nil {
		return "Room already exists - use `\\join` to join the chat.", false
	}
	// Defer first since this since it also locks the mutex and this should run last.
	// Leave any existing rooms this user is in since you can only be in 1.
	defer c.leaveRoom(c.CurrentRoom)

	ChatCache.Mutex.Lock()
	defer ChatCache.Mutex.Unlock()

	log.Printf("Creating Room: %s\n", roomName)
	ChatCache.Rooms[roomName] = []*Client{c}

	c.CurrentRoom = roomName
	return fmt.Sprintf("New room created: %s", roomName), false
}

func (c *Client) joinRoom(roomName string) (string, bool) {
	if ChatCache.Rooms[roomName] == nil {
		return fmt.Sprintf("Room `%s` doesn't exist - try creating it with `\\create`", roomName), false
	}
	if c.CurrentRoom == roomName {
		return fmt.Sprintf("You're already in %s!", roomName), false
	}
	// Defer first since this since it also locks the mutex and this should run last.
	// Leave any existing rooms this user is in since you can only be in 1.
	defer c.leaveRoom(c.CurrentRoom)

	ChatCache.Mutex.Lock()
	defer ChatCache.Mutex.Unlock()

	ChatCache.Rooms[roomName] = append(ChatCache.Rooms[roomName], c)

	c.CurrentRoom = roomName
	return fmt.Sprintf("%s has entered: %s", c.Name, roomName), true
}

func (c *Client) leaveRoom(roomName string) {
	// Return as a no-op here if you didn't give a roomName (you may not have entered a room yet).
	if roomName == "" {
		return
	}

	ChatCache.Mutex.Lock()
	defer ChatCache.Mutex.Unlock()

	// Re-create the list of clients in the given room, minus whatever client we have in play since they are leaving.
	prunedList := []*Client{}
	for _, client := range ChatCache.Rooms[roomName] {
		if c != client {
			prunedList = append(prunedList, client)
		}
	}
	ChatCache.Rooms[roomName] = prunedList

	/// If the room no longer has anyone in it after this user has been removed then delete it.
	if len(ChatCache.Rooms[roomName]) == 0 {
		delete(ChatCache.Rooms, roomName)
	}

	// I hate to do this in here, but I don't really want to pass roomName up through all these methods and
	//their associated conditions when 90% of the time it's going to be what's already on the client.  So
	//leaving this for now.
	go c.broadcastToRoom(fmt.Sprintf("%s has left %s.", c.Name, roomName), roomName)
	// TODO - Consider the map version below relying on this data structure
	//      {<room name>: {<client id>: <client pointer>} }
	//// Delete this client from the rooms: clients mappings.
	//if ChatCache.Rooms[roomName][clientName] != nil {
	//    delete(ChatCache.Rooms[roomName], clientName)
	//}
}

func (c *Client) broadcastToRoom(message, roomName string) {
	// If no one is in the room I'm in then just send it to myself.
	if len(ChatCache.Rooms[roomName]) < 1 {
		c.WriteResponse(message, nil)
	}
	for _, targetClient := range ChatCache.Rooms[roomName] {
		targetClient.WriteResponse(message, c.Name)
	}
}

func (c *Client) listen() {
	r := bufio.NewReader(c.Conn)
	defer c.removeConnection()

	for {
		input, err := Read(r)
		if err != nil && err != io.EOF {
			log.Printf("Read error: %v\n", err)
			c.WriteResponse(input, nil)
		}

		// NOTE: This fires only when the Client kills its connection
		if err == io.EOF {
			break
		}

		if input != "" {
			// These should be commands from the user
			if strings.HasPrefix(input, "\\") {
				response, toBroadcast := c.parseResponse(input)
				if response != "" {
					if toBroadcast {
						go c.broadcastToRoom(response, c.CurrentRoom)
					} else {
						c.WriteResponse(response, nil)
					}
				}
			} else if c.CurrentRoom != "" {
				go c.broadcastToRoom(input, c.CurrentRoom)
			} else {
				c.WriteResponse(input, nil)
			}
		}

	}
}

// Here we will handle any commands and return anything we want to send back to the client.  If we want this to
//  be "broadcast" (to every client in the chat room) we can answer `true` with the accompanying bool.  Most methods
// should actually return their own values to decide whether their message gets broadcast or not, so we can send
//error messages privately, etc.
func (c *Client) parseResponse(cmd string) (string, bool) {
	value := ""
	// Attempt to split the command from the proceeding value, determined by a space (if applicable)
	cmdIndex := strings.IndexByte(cmd, ' ')
	if cmdIndex > 0 {
		value = strings.TrimSpace(cmd[cmdIndex:])
		cmd = cmd[:cmdIndex]
	}
	switch {
	case cmd == "\\name" && value != "":
		return c.changeClientName(value)
	case cmd == "\\create" && value != "":
		return c.createRoom(value)
	case cmd == "\\join" && value != "":
		return c.joinRoom(value)
	case cmd == "\\list" && value != "": // list people in another room
		return c.listMembers(value)
	case cmd == "\\leave":
		roomName := c.CurrentRoom
		c.leaveRoom(c.CurrentRoom)
		c.CurrentRoom = "" // Set this here because leaveRoom is called from all over

		return fmt.Sprintf("You have left room %s", roomName), false
	case cmd == "\\list": // list the members of your current room
		return c.listMembers(c.CurrentRoom)
	case cmd == "\\list-rooms":
		return c.listRooms()
	case cmd == "\\whoami":
		return c.displayClientStats()
	}
	// TODO - \\create-private - private room??? - How would that even work?
	// TODO - Direct message?
	return fmt.Sprintf("Invalid command: `%s`", cmd), false
}
