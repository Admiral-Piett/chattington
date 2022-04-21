package clients

import (
    "chat-telnet/mocks"
    cache2 "github.com/patrickmn/go-cache"
    "github.com/stretchr/testify/assert"
    "testing"
)

func Test_addClientToCache_success_clients_data_found(t *testing.T){
    cm := &mocks.CacheMock{}
    cm.GetMock = func(k string) (interface{}, bool) {
        return map[string]*Client{"456": &Client{}}, true
    }

    c := &Client{Id: "123", Name: "Han Solo", Cache: cm}
    err := c.addClientToCache()

    assert.Nil(t, err)
    assert.Equal(t, CLIENTS, cm.GetCalledWithKey)
    assert.Equal(t, CLIENTS, cm.SetCalledWithKey)
    assert.Equal(t, map[string]*Client{"456": &Client{}, c.Id: c}, cm.SetCalledWithInterface)
    assert.Equal(t, cache2.NoExpiration, cm.SetCalledWithDuration)
}

func Test_addClientToCache_success_no_clients_data_found(t *testing.T){
    cm := &mocks.CacheMock{}

    c := &Client{Id: "123", Name: "Han Solo", Cache: cm}
    err := c.addClientToCache()

    assert.Nil(t, err)
    assert.Equal(t, CLIENTS, cm.GetCalledWithKey)
    assert.Equal(t, CLIENTS, cm.SetCalledWithKey)
    assert.Equal(t, map[string]*Client{c.Id: c}, cm.SetCalledWithInterface)
    assert.Equal(t, cache2.NoExpiration, cm.SetCalledWithDuration)
}

func Test_addClientToCache_success_client_already_in_cache(t *testing.T){
    cm := &mocks.CacheMock{}
    c := &Client{Id: "123", Name: "Han Solo", Cache: cm}

    cm.GetMock = func(k string) (interface{}, bool) {
        return map[string]*Client{c.Id: c}, true
    }
    err := c.addClientToCache()

    assert.Error(t, err)
    assert.Equal(t, CLIENTS, cm.GetCalledWithKey)
    assert.False(t, cm.SetCalled)
}

func Test_removeClientFromCache_success_clients_data_found(t *testing.T){
    cm := &mocks.CacheMock{}
    c := &Client{Id: "123", Name: "Han Solo", Cache: cm}

    cm.GetMock = func(k string) (interface{}, bool) {
        return map[string]*Client{c.Id: c}, true
    }
    c.removeClientFromCache()

    assert.Equal(t, CLIENTS, cm.GetCalledWithKey)
    assert.Equal(t, CLIENTS, cm.SetCalledWithKey)
    assert.Equal(t, map[string]*Client{}, cm.SetCalledWithInterface)
    assert.Equal(t, cache2.NoExpiration, cm.SetCalledWithDuration)
}

func Test_removeClientFromCache_success_no_clients_data_found(t *testing.T){
    cm := &mocks.CacheMock{}
    c := &Client{Id: "123", Name: "Han Solo", Cache: cm}

    c.removeClientFromCache()

    assert.Equal(t, CLIENTS, cm.GetCalledWithKey)
    assert.Equal(t, CLIENTS, cm.SetCalledWithKey)
    assert.Equal(t, map[string]*Client{}, cm.SetCalledWithInterface)
    assert.Equal(t, cache2.NoExpiration, cm.SetCalledWithDuration)
}

func Test_updateClientInCache_success_clients_data_found_add_new(t *testing.T){
    cm := &mocks.CacheMock{}
    c := &Client{Id: "123", Name: "Han Solo", Cache: cm}

    cm.GetMock = func(k string) (interface{}, bool) {
        return map[string]*Client{"456": &Client{}}, true
    }
    c.updateClientInCache()

    assert.Equal(t, CLIENTS, cm.GetCalledWithKey)
    assert.Equal(t, CLIENTS, cm.SetCalledWithKey)
    assert.Equal(t, map[string]*Client{"456": &Client{}, c.Id: c}, cm.SetCalledWithInterface)
    assert.Equal(t, cache2.NoExpiration, cm.SetCalledWithDuration)
}

func Test_updateClientInCache_success_clients_data_found_overwrite_on_matching_ids(t *testing.T){
    cm := &mocks.CacheMock{}
    c := &Client{Id: "123", Name: "Han Solo", Cache: cm}

    cm.GetMock = func(k string) (interface{}, bool) {
        return map[string]*Client{c.Id: c}, true
    }
    c2 := &Client{Id: "123", Name: "Lando Calrissian", Cache: cm}
    c2.updateClientInCache()

    assert.Equal(t, CLIENTS, cm.GetCalledWithKey)
    assert.Equal(t, CLIENTS, cm.SetCalledWithKey)
    assert.Equal(t, map[string]*Client{c2.Id: c2}, cm.SetCalledWithInterface)
    assert.Equal(t, cache2.NoExpiration, cm.SetCalledWithDuration)
}

func Test_updateClientInCache_success_no_clients_data_found(t *testing.T){
    cm := &mocks.CacheMock{}
    c := &Client{Id: "123", Name: "Han Solo", Cache: cm}

    c.updateClientInCache()

    assert.Equal(t, CLIENTS, cm.GetCalledWithKey)
    assert.Equal(t, CLIENTS, cm.SetCalledWithKey)
    assert.Equal(t, map[string]*Client{c.Id: c}, cm.SetCalledWithInterface)
    assert.Equal(t, cache2.NoExpiration, cm.SetCalledWithDuration)
}

func Test_getAllRoomsFromCache_success_rooms_data_found(t *testing.T){
    cm := &mocks.CacheMock{}
    c := &Client{Id: "123", Name: "Han Solo", Cache: cm}
    cm.GetMock = func(k string) (interface{}, bool) {
        return map[string][]*Client{"broom": {c}}, true
    }

    results := c.getAllRoomsFromCache()

    assert.Equal(t, ROOMS, cm.GetCalledWithKey)
    assert.Equal(t, map[string][]*Client{"broom": {c}}, results)
}

func Test_getAllRoomsFromCache_success_no_rooms_data_found(t *testing.T){
    cm := &mocks.CacheMock{}
    c := &Client{Id: "123", Name: "Han Solo", Cache: cm}

    results := c.getAllRoomsFromCache()

    assert.Equal(t, ROOMS, cm.GetCalledWithKey)
    assert.Equal(t, map[string][]*Client{}, results)
}

func Test_getRoomFromCacheByName_success_single_room(t *testing.T){
    cm := &mocks.CacheMock{}
    c := &Client{Id: "123", Name: "Han Solo", Cache: cm}
    cm.GetMock = func(k string) (interface{}, bool) {
        return map[string][]*Client{"broom": {c}}, true
    }

    results, found := c.getRoomFromCacheByName("broom")

    assert.Equal(t, ROOMS, cm.GetCalledWithKey)
    assert.True(t, found)
    assert.Equal(t, []*Client{c}, results)
}

func Test_getRoomFromCacheByName_success_multiple_rooms(t *testing.T){
    cm := &mocks.CacheMock{}
    c := &Client{Id: "123", Name: "Han Solo", Cache: cm}
    cm.GetMock = func(k string) (interface{}, bool) {
        return map[string][]*Client{"broom": {c}, "vroom": {&Client{}}}, true
    }

    results, found := c.getRoomFromCacheByName("vroom")

    assert.Equal(t, ROOMS, cm.GetCalledWithKey)
    assert.True(t, found)
    assert.Equal(t, []*Client{&Client{}}, results)
}

func Test_getRoomFromCacheByName_success_no_rooms_data_found(t *testing.T){
    cm := &mocks.CacheMock{}
    c := &Client{Id: "123", Name: "Han Solo", Cache: cm}

    results, found := c.getRoomFromCacheByName("broom")

    assert.Equal(t, ROOMS, cm.GetCalledWithKey)
    assert.False(t, found)
    assert.Nil(t, results)
}

func Test_getRoomFromCacheByName_success_invalid_room_name(t *testing.T){
    cm := &mocks.CacheMock{}
    c := &Client{Id: "123", Name: "Han Solo", Cache: cm}
    cm.GetMock = func(k string) (interface{}, bool) {
        return map[string][]*Client{"broom": {c}}, true
    }

    results, found := c.getRoomFromCacheByName("blah")

    assert.Equal(t, ROOMS, cm.GetCalledWithKey)
    assert.False(t, found)
    assert.Nil(t, results)
}

func Test_updateRoomInCache_success_rooms_data_found_add_new(t *testing.T){
    cm := &mocks.CacheMock{}
    c := &Client{Id: "123", Name: "Han Solo", Cache: cm}
    cm.GetMock = func(k string) (interface{}, bool) {
        return map[string][]*Client{"broom": {c}}, true
    }

    c.updateRoomInCache("vroom", []*Client{&Client{}})

    assert.Equal(t, ROOMS, cm.GetCalledWithKey)
    assert.Equal(t, ROOMS, cm.SetCalledWithKey)
    assert.Equal(t, map[string][]*Client{"broom": {c}, "vroom": {&Client{}}}, cm.SetCalledWithInterface)
    assert.Equal(t, cache2.NoExpiration, cm.SetCalledWithDuration)
}

func Test_updateRoomInCache_success_rooms_data_found_overwrite_on_matching_ids(t *testing.T){
    cm := &mocks.CacheMock{}
    c := &Client{Id: "123", Name: "Han Solo", Cache: cm}
    cm.GetMock = func(k string) (interface{}, bool) {
        return map[string][]*Client{"broom": {&Client{}}}, true
    }

    c.updateRoomInCache("broom", []*Client{c})

    assert.Equal(t, ROOMS, cm.GetCalledWithKey)
    assert.Equal(t, ROOMS, cm.SetCalledWithKey)
    assert.Equal(t, map[string][]*Client{"broom": {c}}, cm.SetCalledWithInterface)
    assert.Equal(t, cache2.NoExpiration, cm.SetCalledWithDuration)
}

func Test_updateRoomInCache_success_no_rooms_data_found(t *testing.T){
    cm := &mocks.CacheMock{}
    c := &Client{Id: "123", Name: "Han Solo", Cache: cm}

    c.updateRoomInCache("broom", []*Client{c})

    assert.Equal(t, ROOMS, cm.GetCalledWithKey)
    assert.Equal(t, ROOMS, cm.SetCalledWithKey)
    assert.Equal(t, map[string][]*Client{"broom": {c}}, cm.SetCalledWithInterface)
    assert.Equal(t, cache2.NoExpiration, cm.SetCalledWithDuration)
}

func Test_deleteRoomFromCache_success_rooms_single_room(t *testing.T){
    cm := &mocks.CacheMock{}
    c := &Client{Id: "123", Name: "Han Solo", Cache: cm}
    cm.GetMock = func(k string) (interface{}, bool) {
        return map[string][]*Client{"broom": {c}}, true
    }

    c.deleteRoomFromCache("broom")

    assert.Equal(t, ROOMS, cm.GetCalledWithKey)
    assert.Equal(t, ROOMS, cm.SetCalledWithKey)
    assert.Equal(t, map[string][]*Client{}, cm.SetCalledWithInterface)
    assert.Equal(t, cache2.NoExpiration, cm.SetCalledWithDuration)
}

func Test_deleteRoomFromCache_success_rooms_multiple_rooms(t *testing.T){
    cm := &mocks.CacheMock{}
    c := &Client{Id: "123", Name: "Han Solo", Cache: cm}
    cm.GetMock = func(k string) (interface{}, bool) {
        return map[string][]*Client{"broom": {&Client{}}, "vroom": {c}}, true
    }

    c.deleteRoomFromCache("vroom")

    assert.Equal(t, ROOMS, cm.GetCalledWithKey)
    assert.Equal(t, ROOMS, cm.SetCalledWithKey)
    assert.Equal(t, map[string][]*Client{"broom": {&Client{}}}, cm.SetCalledWithInterface)
    assert.Equal(t, cache2.NoExpiration, cm.SetCalledWithDuration)
}

func Test_deleteRoomFromCache_success_no_rooms_data_found(t *testing.T){
    cm := &mocks.CacheMock{}
    c := &Client{Id: "123", Name: "Han Solo", Cache: cm}

    c.deleteRoomFromCache("vroom")

    assert.Equal(t, ROOMS, cm.GetCalledWithKey)
    assert.Equal(t, ROOMS, cm.SetCalledWithKey)
    assert.Equal(t, map[string][]*Client{}, cm.SetCalledWithInterface)
    assert.Equal(t, cache2.NoExpiration, cm.SetCalledWithDuration)
}
