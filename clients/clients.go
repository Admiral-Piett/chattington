package clients

import (
	"bufio"
	"fmt"
	"github.com/Admiral-Piett/chat-telnet/interfaces"
	cache2 "github.com/patrickmn/go-cache"
	"io"
	"log"
	"strings"
	"time"
)

var CLIENTS = "clients"
var ROOMS = "rooms"

type Client struct {
	Writer      interfaces.AbstractIoWriter
	Conn        interfaces.AbstractNetConn
	Cache       interfaces.AbstractCache // TODO - explore a cache like redis or BadgerDB?
	Name        string
	CurrentRoom string
	Id          string
}

func GenerateNewClient(conn interfaces.AbstractNetConn, cache interfaces.AbstractCache) error {
	log.Printf("Accepting new connection from address %v\n", conn.RemoteAddr().String())

	// Generate a semi-random id and initial name for ourselves, using a time stamp.  For now, this is "unique"
	//  enough since our users are not logging on so quickly that they should collide.
	id := fmt.Sprintf("%v", time.Now().Unix())
	client := &Client{
		Conn:        conn,
		Writer:      conn,
		Name:        id,
		CurrentRoom: "",
		Id:          id,
		Cache:       cache,
	}

	err := client.addClientToCache()
	if err != nil {
		client.WriteString(fmt.Sprintf("ERROR: %s\n", err))
		return err
	}
	intro := `
Welcome to Chattington!

Feel free to join any chat rooms you see, or create a room instead, using the available commands below.

Available Commands:
=====
\name 	<user name> 	: Change your user name to the <user name> supplied
\create <room name> 	: Create and join a new chat room with the <room name> supplied
\join 	<room name> 	: Join an existing chat room with the <room name> supplied
\list 	<room name>		: List members in the chat room named after the <room name> supplied
\leave					: Leave the room you are currently in
\list 					: List members in the room you're currently in
\list-rooms				: List all the available rooms and their members
\whoami					: List your name and what room you're currently in
\exit					: Exit server and terminate connection
`
	nameInstructions := fmt.Sprintf("\n\nNOTE: Your user name has been automatically set to `%s`\nIf you'd like to reset it, please use the '\\name' command.\n\n", id)

	client.WriteString(intro + nameInstructions)
	go client.listen()
	return nil
}

func (c *Client) WriteString(msg string) error {
	_, err := c.Writer.Write([]byte(msg))

	return err
}

func (c *Client) WriteResponse(msg string, sendingClient interface{}) error {
	// Using `sendingClient` you can send in the string name who the sender of the message is, if it's `nil` we'll
	//  or the same as the target client (`c`) then we can format the response as i.
	now := time.Now().Unix()
	prefix := ""
	if sendingClient == nil || sendingClient == c.Name {
		prefix = fmt.Sprintf("%d: %s>", now, c.Name)
	} else {
		prefix = fmt.Sprintf("%d: %s:", now, sendingClient)
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
	c.removeClientFromCache()
	c.Conn.Close()
	log.Printf("Removed connection %v from pool\n", c.Conn.RemoteAddr().String())
}

func (c *Client) changeClientName(name string) (string, bool) {
	msg := fmt.Sprintf("User: %s has become -> %s", c.Name, name)

	c.Name = name
	c.updateClientInCache()
	// Overwrite this user in the cache so all users can see it.
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
	rooms := c.getAllRoomsFromCache()
	if len(rooms) < 1 {
		return "No rooms yet - make one!", false
	}
	roomString := ""
	for name, members := range rooms {
		roomString = roomString + fmt.Sprintf("  Room: %s\n  Members:\n", name)
		for _, c := range members {
			roomString = roomString + fmt.Sprintf("\t%s\n", c.Name)
		}
	}
	return fmt.Sprintf("\nCurrent rooms: \n%s", roomString), false
}

func (c *Client) listMembers(roomName string) (string, bool) {
	room, found := c.getRoomFromCacheByName(roomName)
	if !found {
		return fmt.Sprintf("No such room %s!", roomName), false
	}
	roomString := ""
	for _, c := range room {
		roomString = roomString + fmt.Sprintf("\t%s\n", c.Name)
	}
	return fmt.Sprintf("\nCurrent Members:\n%s", roomString), false
}

func (c *Client) createRoom(roomName string) (string, bool) {
	_, found := c.getRoomFromCacheByName(roomName)
	if found {
		return "Room already exists - use `\\join` to join the chat.", false
	}
	// Defer first since this since it also locks the mutex and this should run last.
	// Leave any existing rooms this user is in since you can only be in 1.
	defer c.leaveRoom(c.CurrentRoom)

	log.Printf("Creating Room: %s\n", roomName)
	c.CurrentRoom = roomName
	c.updateRoomInCache(roomName, []*Client{c})

	return fmt.Sprintf("New room created: %s", roomName), false
}

func (c *Client) joinRoom(roomName string) (string, bool) {
	room, found := c.getRoomFromCacheByName(roomName)
	if !found {
		return fmt.Sprintf("Room `%s` doesn't exist - try creating it with `\\create`", roomName), false
	}
	if c.CurrentRoom == roomName {
		return fmt.Sprintf("You're already in %s!", roomName), false
	}
	// Defer first since this since it also locks the mutex and this should run last.
	// Leave any existing rooms this user is in since you can only be in 1.
	defer c.leaveRoom(c.CurrentRoom)

	c.CurrentRoom = roomName

	room = append(room, c)
	c.updateRoomInCache(roomName, room)
	return fmt.Sprintf("%s has entered: %s", c.Name, roomName), true
}

func (c *Client) leaveRoom(roomName string) {
	// Return as a no-op here if you didn't give a roomName (you may not have entered a room yet).
	if roomName == "" {
		return
	}

	room, found := c.getRoomFromCacheByName(roomName)
	// If this room doesn't exist then just return as a no-op here
	if !found {
		return
	}
	// Re-create the list of clients in the given room, minus whatever client we have in play since they are leaving.
	prunedList := []*Client{}
	for _, client := range room {
		if c != client {
			prunedList = append(prunedList, client)
		}
	}

	/// If the room no longer has anyone in it after this user has been removed then delete it.
	if len(prunedList) == 0 {
		c.deleteRoomFromCache(roomName)
	} else {
		c.updateRoomInCache(roomName, prunedList)
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

func (c *Client) directMessage(targetClient string) (string, bool) {
	return "", false
}

func (c *Client) broadcastToRoom(message, roomName string) {
	// We don't care if the room was found or not, since we'll detect and empty room (or one where this client is
	//the only one in it) and send the message only to that client.
	room, _ := c.getRoomFromCacheByName(roomName)
	// If no one is in the room I'm in then just send it to myself.
	if len(room) < 1 {
		c.WriteResponse(message, nil)
	}
	for _, targetClient := range room {
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
				response, toBroadcast, err := c.parseResponse(input)
				if err != nil {
					go c.broadcastToRoom(response, c.CurrentRoom)
					break // Sever the connection to this client
				}
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
// TODO - create response type structs instead of returning all this willy nilly?
func (c *Client) parseResponse(cmd string) (string, bool, error) {
	value := ""
	// Attempt to split the command from the proceeding value, determined by a space (if applicable)
	cmdIndex := strings.IndexByte(cmd, ' ')
	if cmdIndex > 0 {
		value = strings.TrimSpace(cmd[cmdIndex:])
		cmd = cmd[:cmdIndex]
	}
	switch {
	case cmd == "\\dm" && value != "":
		response, toBroadcast := c.directMessage(value) //TODO - work this out
		return response, toBroadcast, nil
	case cmd == "\\name" && value != "":
		response, toBroadcast := c.changeClientName(value)
		return response, toBroadcast, nil
	case cmd == "\\create" && value != "":
		response, toBroadcast := c.createRoom(value)
		return response, toBroadcast, nil
	case cmd == "\\join" && value != "":
		response, toBroadcast := c.joinRoom(value)
		return response, toBroadcast, nil
	case cmd == "\\list" && value != "": // list people in another room
		response, toBroadcast := c.listMembers(value)
		return response, toBroadcast, nil
	case cmd == "\\leave":
		roomName := c.CurrentRoom
		c.leaveRoom(c.CurrentRoom)
		c.CurrentRoom = "" // Set this here because leaveRoom is called from all over

		return fmt.Sprintf("You have left room %s", roomName), false, nil
	case cmd == "\\list": // list the members of your current room
		response, toBroadcast := c.listMembers(c.CurrentRoom)
		return response, toBroadcast, nil
	case cmd == "\\list-rooms":
		response, toBroadcast := c.listRooms()
		return response, toBroadcast, nil
	case cmd == "\\whoami":
		response, toBroadcast := c.displayClientStats()
		return response, toBroadcast, nil
	case cmd == "\\cancel": // cancel current activity (\dm-ing)
		return "", false, nil //TODO - work this out
	case cmd == "\\exit":
		return fmt.Sprintf("%s has gone offline", c.Name), true, io.EOF
	}
	return fmt.Sprintf("Invalid command: `%s`", cmd), false, nil
}

func (c *Client) addClientToCache() error {
	cc := map[string]*Client{}
	chatClients, found := c.Cache.Get(CLIENTS)
	if found {
		cc = chatClients.(map[string]*Client)
	}
	// If somehow we already have this client in the cache return and error
	if cc[c.Id] != nil {
		return fmt.Errorf("User Conflict: %s user already in service. Please try again.", c.Name)
	}
	cc[c.Id] = c
	c.Cache.Set(CLIENTS, cc, cache2.NoExpiration)
	return nil
}

func (c *Client) removeClientFromCache() {
	cc := map[string]*Client{}
	chatClients, found := c.Cache.Get(CLIENTS)
	if found {
		cc = chatClients.(map[string]*Client)
	}
	delete(cc, c.Id)
	c.Cache.Set(CLIENTS, cc, cache2.NoExpiration)
}

// NOTE: This and other update methods are not always strictly necessary because we are operating on *Client objects.
//The pointer is carrying the updates into the memory cache.  Leaving this here though as a good DB pattern, should
//I decide to update to a more advanced cache or DB beyond this.
func (c *Client) updateClientInCache() {
	cc := map[string]*Client{}
	chatClients, found := c.Cache.Get(CLIENTS)
	if found {
		cc = chatClients.(map[string]*Client)
	}
	cc[c.Id] = c
	c.Cache.Set(CLIENTS, cc, cache2.NoExpiration)
}

func (c *Client) getAllRoomsFromCache() map[string][]*Client {
	rc := map[string][]*Client{}
	rooms, found := c.Cache.Get(ROOMS)
	if found {
		rc = rooms.(map[string][]*Client)
	}
	return rc
}

// Return the chat room, containing pointers to all the clients currently in the room and a bool indicating whether
//	or not the room already exists
func (c *Client) getRoomFromCacheByName(roomName string) ([]*Client, bool) {
	rc := map[string][]*Client{}
	rooms, found := c.Cache.Get(ROOMS)
	if found {
		rc = rooms.(map[string][]*Client)
	}
	if rc[roomName] != nil {
		return rc[roomName], true
	}
	return rc[roomName], false
}

func (c *Client) updateRoomInCache(roomName string, clientList []*Client) {
	rc := map[string][]*Client{}
	rooms, found := c.Cache.Get(ROOMS)
	if found {
		rc = rooms.(map[string][]*Client)
	}
	rc[roomName] = clientList
	c.Cache.Set(ROOMS, rc, cache2.NoExpiration)
}

func (c *Client) deleteRoomFromCache(roomName string) {
	rc := map[string][]*Client{}
	rooms, found := c.Cache.Get(ROOMS)
	if found {
		rc = rooms.(map[string][]*Client)
	}
	delete(rc, roomName)
	c.Cache.Set(ROOMS, rc, cache2.NoExpiration)
}
