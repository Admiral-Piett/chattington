package clients

import (
    "fmt"
    cache2 "github.com/patrickmn/go-cache"
)

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
