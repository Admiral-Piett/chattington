package clients

import (
	"bou.ke/monkey"
	"chat-telnet/interfaces"
	"chat-telnet/mocks"
	"fmt"
	cache2 "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func Test_GenerateNewClient_success(t *testing.T) {
	// Kill the listen loop
	monkey.Patch(Read, func(a interfaces.AbstractBufioReader) (string, error) {
		return "", io.EOF
	})
	defer monkey.Unpatch(Read)
	err := GenerateNewClient(&mocks.NetConnMock{}, cache2.New(cache2.NoExpiration, cache2.NoExpiration))
	assert.Nil(t, err)
}

func Test_GenerateNewClient_client_already_exists_error(t *testing.T) {
	// Kill the listen loop
	monkey.Patch(Read, func(a interfaces.AbstractBufioReader) (string, error) {
		return "", io.EOF
	})
	defer monkey.Unpatch(Read)
	monkey.Patch(time.Now, func() time.Time {
		return time.Date(2022, 04, 20, 11, 00, 00, 00, time.UTC)
	})
	defer monkey.Unpatch(time.Now)

	cm := &mocks.CacheMock{}
	cm.GetMock = func(k string) (interface{}, bool) {
		return map[string]*Client{"1650452400": &Client{}}, true
	}

	err := GenerateNewClient(&mocks.NetConnMock{}, cm)
	assert.Equal(t, "User Conflict: 1650452400 user already in service. Please try again.", fmt.Sprint(err))
}

func Test_WriteString_success(t *testing.T) {
	w := &mocks.IoWriterMock{}
	m := &Client{
		Writer:      w,
		Conn:        &mocks.NetConnMock{},
		Name:        "Han Solo",
		CurrentRoom: "",
		Id:          "test-id",
	}
	err := m.WriteString("Hi")

	assert.Nil(t, err)
	assert.Equal(t, "Hi", string(w.WriteCalledWith))
}

func Test_WriteString_write_error(t *testing.T) {
	w := &mocks.IoWriterMock{
		WriteMock: func(p []byte) (n int, err error) {
			return 0, fmt.Errorf("boom")
		},
	}
	m := &Client{
		Writer:      w,
		Conn:        &mocks.NetConnMock{},
		Name:        "Han Solo",
		CurrentRoom: "",
		Id:          "test-id",
	}
	err := m.WriteString("Hi")

	assert.Error(t, err)
}

func Test_WriteResponse_success_no_sendingClient(t *testing.T) {
	w := &mocks.IoWriterMock{}
	m := &Client{
		Writer:      w,
		Conn:        &mocks.NetConnMock{},
		Name:        "Han Solo",
		CurrentRoom: "",
		Id:          "test-id",
	}
	monkey.Patch(time.Now, func() time.Time {
		return time.Date(2022, 04, 20, 11, 00, 00, 00, time.UTC)
	})
	defer monkey.Unpatch(time.Now)

	err := m.WriteResponse("Hi", nil)

	assert.Nil(t, err)
	assert.Equal(t, "1650452400: Han Solo> Hi\n", string(w.WriteCalledWith))
}

func Test_WriteResponse_success_with_sendingClient(t *testing.T) {
	w := &mocks.IoWriterMock{}
	m := &Client{
		Writer:      w,
		Conn:        &mocks.NetConnMock{},
		Name:        "Han Solo",
		CurrentRoom: "",
		Id:          "test-id",
	}
	monkey.Patch(time.Now, func() time.Time {
		return time.Date(2022, 04, 20, 11, 00, 00, 00, time.UTC)
	})
	defer monkey.Unpatch(time.Now)
	err := m.WriteResponse("Hi", "Leia Organa")

	assert.Nil(t, err)
	assert.Equal(t, "1650452400: Leia Organa: Hi\n", string(w.WriteCalledWith))
}

func Test_WriteResponse_returns_error(t *testing.T) {
	w := &mocks.IoWriterMock{
		WriteMock: func(p []byte) (n int, err error) {
			return 0, fmt.Errorf("boom")
		},
	}
	m := &Client{
		Writer:      w,
		Conn:        &mocks.NetConnMock{},
		Name:        "Han Solo",
		CurrentRoom: "",
		Id:          "test-id",
	}
	err := m.WriteResponse("Hi", nil)

	assert.Error(t, err)
}

func Test_Read_success(t *testing.T) {
	m := &mocks.ReaderMock{}
	_, err := Read(m)

	assert.Nil(t, err)
	assert.True(t, m.ReadStringCalled)
}

func Test_Read_returns_error(t *testing.T) {
	m := &mocks.ReaderMock{
		ReadMock: func(delim byte) (string, error) {
			return "", fmt.Errorf("boom")
		},
	}
	s, err := Read(m)

	assert.Error(t, err)
	assert.Empty(t, s)
	assert.True(t, m.ReadStringCalled)
}

func Test_removeConnection_success(t *testing.T) {
	cm := &mocks.CacheMock{}
	c := &Client{Id: "123", Conn: &mocks.NetConnMock{}, Cache: cm}

	// Seed the cache
	cm.GetMock = func(k string) (interface{}, bool) {
		return map[string]*Client{c.Id: c}, true
	}

	c.removeConnection()

	assert.Equal(t, CLIENTS, cm.GetCalledWithKey)
	assert.Equal(t, CLIENTS, cm.SetCalledWithKey)
	assert.Equal(t, map[string]*Client{}, cm.SetCalledWithInterface)
	assert.Equal(t, cache2.NoExpiration, cm.SetCalledWithDuration)
}

func Test_changeClientName_success(t *testing.T) {
	cm := &mocks.CacheMock{}
	c := &Client{Id: "123", Name: "Han Solo", Cache: cm}
	cm.GetMock = func(k string) (interface{}, bool) {
		return map[string]*Client{c.Id: c}, true
	}
	c.changeClientName("Luke Skywalker")

	assert.Equal(t, CLIENTS, cm.GetCalledWithKey)
	assert.Equal(t, CLIENTS, cm.SetCalledWithKey)
	assert.IsType(t, map[string]*Client{}, cm.SetCalledWithInterface)
	assert.Equal(t, "Luke Skywalker", cm.SetCalledWithInterface.(map[string]*Client)[c.Id].Name)
	assert.Equal(t, cache2.NoExpiration, cm.SetCalledWithDuration)
}

func Test_displayClientStats_success(t *testing.T) {
	c := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom"}
	response, b := c.displayClientStats()

	assert.Equal(t, "\nClient Name: Han Solo\nCurrent Room: broom", response)
	assert.False(t, b)
}

func Test_listRooms_success(t *testing.T) {
	cm := &mocks.CacheMock{}
	c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Cache: cm}
	c2 := &Client{Id: "456", Name: "Chewbacca", CurrentRoom: "broom", Cache: cm}
	cm.GetMock = func(k string) (interface{}, bool) {
		return map[string][]*Client{"broom": {c1, c2}}, true
	}
	response, b := c1.listRooms()

	assert.Equal(t, "\nCurrent rooms: \n  Room: broom\n  Members:\n\tHan Solo\n\tChewbacca\n", response)
	assert.False(t, b)

	assert.Equal(t, ROOMS, cm.GetCalledWithKey)
	assert.False(t, cm.SetCalled)
}

func Test_listRooms_no_rooms(t *testing.T) {
	cm := &mocks.CacheMock{}
	c := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Cache: cm}
	cm.GetMock = func(k string) (interface{}, bool) {
		return map[string][]*Client{}, false
	}

	response, b := c.listRooms()

	assert.Equal(t, "No rooms yet - make one!", response)
	assert.False(t, b)
	assert.Equal(t, ROOMS, cm.GetCalledWithKey)
}

func Test_listMembers_success(t *testing.T) {
	cm := &mocks.CacheMock{}
	c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Cache: cm}
	c2 := &Client{Id: "456", Name: "Chewbacca", CurrentRoom: "broom", Cache: cm}
	cm.GetMock = func(k string) (interface{}, bool) {
		return map[string][]*Client{"broom": {c1, c2}}, true
	}
	response, b := c1.listMembers("broom")

	assert.Equal(t, "\nCurrent Members:\n\tHan Solo\n\tChewbacca\n", response)
	assert.False(t, b)
	assert.Equal(t, ROOMS, cm.GetCalledWithKey)
}

func Test_listMembers_invalid_roomName(t *testing.T) {
	cm := &mocks.CacheMock{}
	c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Cache: cm}
	c2 := &Client{Id: "456", Name: "Chewbacca", CurrentRoom: "broom", Cache: cm}
	cm.GetMock = func(k string) (interface{}, bool) {
		return map[string][]*Client{"broom": {c1, c2}}, true
	}
	response, b := c1.listMembers("vroom")

	assert.Equal(t, "No such room vroom!", response)
	assert.False(t, b)
	assert.Equal(t, ROOMS, cm.GetCalledWithKey)
}

func Test_createRoom_success(t *testing.T) {
	cm := &mocks.CacheMock{}
	c := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: &mocks.IoWriterMock{}, Cache: cm}
	cm.GetMock = func(k string) (interface{}, bool) {
		return map[string][]*Client{"broom": {c}}, true
	}

	response, b := c.createRoom("mushroom")

	assert.Equal(t, "New room created: mushroom", response)
	assert.False(t, b)

	// TODO - figure out a good pattern to assert on function taking multiple calls.  Right now my mock will only
	// 	provide the calls for the last one.
	assert.Equal(t, ROOMS, cm.GetCalledWithKey)
	assert.Equal(t, 2, cm.SetCallCount)
}

func Test_joinRoom_success(t *testing.T) {
	cm := &mocks.CacheMock{}
	c := &Client{Id: "123", Name: "Han Solo", Writer: &mocks.IoWriterMock{}, Cache: cm}
	cm.GetMock = func(k string) (interface{}, bool) {
		return map[string][]*Client{"broom": {}}, true
	}

	response, b := c.joinRoom("broom")

	assert.Equal(t, "Han Solo has entered: broom", response)
	assert.True(t, b)

	assert.Equal(t, "broom", c.CurrentRoom)
	assert.Equal(t, ROOMS, cm.GetCalledWithKey)
	assert.Equal(t, ROOMS, cm.SetCalledWithKey)
	assert.Equal(t, []*Client{c}, cm.SetCalledWithInterface.(map[string][]*Client)["broom"])
}

func Test_joinRoom_room_does_not_exist(t *testing.T) {
	cm := &mocks.CacheMock{}
	c := &Client{Id: "123", Name: "Han Solo", Writer: &mocks.IoWriterMock{}, Cache: cm}
	cm.GetMock = func(k string) (interface{}, bool) {
		return map[string][]*Client{}, true
	}

	response, b := c.joinRoom("vroom")

	assert.Equal(t, "", c.CurrentRoom)
	assert.Equal(t, "Room `vroom` doesn't exist - try creating it with `\\create`", response)
	assert.False(t, b)

	assert.False(t, cm.SetCalled)
}

func Test_joinRoom_already_in_room(t *testing.T) {
	cm := &mocks.CacheMock{}
	c := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: &mocks.IoWriterMock{}, Cache: cm}
	cm.GetMock = func(k string) (interface{}, bool) {
		return map[string][]*Client{"broom": {c}}, true
	}
	response, b := c.joinRoom("broom")

	assert.Equal(t, "broom", c.CurrentRoom)
	assert.Equal(t, "You're already in broom!", response)
	assert.False(t, b)

	assert.False(t, cm.SetCalled)
}

func Test_leaveRoom_success(t *testing.T) {
	cm := &mocks.CacheMock{}
	w2 := &mocks.IoWriterMock{}
	c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: &mocks.IoWriterMock{}, Cache: cm}
	c2 := &Client{Id: "456", Name: "Chewbacca", CurrentRoom: "broom", Writer: w2, Cache: cm}
	cm.GetMock = func(k string) (interface{}, bool) {
		return map[string][]*Client{"broom": {c1, c2}}, true
	}

	c1.leaveRoom("broom")

	assert.Equal(t, ROOMS, cm.GetCalledWithKey)
	assert.Equal(t, ROOMS, cm.SetCalledWithKey)
	assert.Equal(t, []*Client{c2}, cm.SetCalledWithInterface.(map[string][]*Client)["broom"])
}

func Test_leaveRoom_empty_string_room_name(t *testing.T) {
	cm := &mocks.CacheMock{}
	w2 := &mocks.IoWriterMock{}
	c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: &mocks.IoWriterMock{}, Cache: cm}
	c2 := &Client{Id: "456", Name: "Chewbacca", CurrentRoom: "broom", Writer: w2, Cache: cm}
	cm.GetMock = func(k string) (interface{}, bool) {
		return map[string][]*Client{"broom": {c1, c2}}, true
	}

	c1.leaveRoom("")

	assert.False(t, w2.WriteCalled)
	assert.False(t, cm.SetCalled)
}

func Test_leaveRoom_empties_out_room_destroys_room(t *testing.T) {
	cm := &mocks.CacheMock{}
	c := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: &mocks.IoWriterMock{}, Cache: cm}
	cm.GetMock = func(k string) (interface{}, bool) {
		return map[string][]*Client{"broom": {c}}, true
	}

	c.leaveRoom("broom")

	assert.Equal(t, map[string][]*Client{}, cm.SetCalledWithInterface.(map[string][]*Client))
}

func Test_broadcastToRoom_success(t *testing.T) {
	cm := &mocks.CacheMock{}
	w1 := &mocks.IoWriterMock{}
	w2 := &mocks.IoWriterMock{}
	c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: w1, Cache: cm}
	c2 := &Client{Id: "456", Name: "Chewbacca", CurrentRoom: "broom", Writer: w2, Cache: cm}
	cm.GetMock = func(k string) (interface{}, bool) {
		return map[string][]*Client{"broom": {c1, c2}}, true
	}

	monkey.Patch(time.Now, func() time.Time {
		return time.Date(2022, 04, 20, 11, 00, 00, 00, time.UTC)
	})
	defer monkey.Unpatch(time.Now)

	c1.broadcastToRoom("test", "broom")

	assert.Equal(t, "1650452400: Han Solo> test\n", string(w1.WriteCalledWith))
	assert.Equal(t, "1650452400: Han Solo: test\n", string(w2.WriteCalledWith))
}

func Test_broadcastToRoom_alone_write_to_self(t *testing.T) {
	cm := &mocks.CacheMock{}
	w1 := &mocks.IoWriterMock{}
	c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: w1, Cache: cm}
	cm.GetMock = func(k string) (interface{}, bool) {
		return map[string][]*Client{"broom": {c1}}, true
	}
	monkey.Patch(time.Now, func() time.Time {
		return time.Date(2022, 04, 20, 11, 00, 00, 00, time.UTC)
	})
	defer monkey.Unpatch(time.Now)

	c1.broadcastToRoom("test", "broom")

	assert.Equal(t, "1650452400: Han Solo> test\n", string(w1.WriteCalledWith))
}

func Test_parseResponse_one_client_required(t *testing.T) {
	cm := cache2.New(cache2.NoExpiration, cache2.NoExpiration)
	c := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "", Writer: &mocks.IoWriterMock{}, Cache: cm}

	var tests = []struct {
		input         string
		expectedStr   string
		expectedBool  bool
		expectedError error
	}{
		{"\\name Lando Calrissian", "User: Han Solo has become -> Lando Calrissian", true, nil},
		{"\\create broom", "New room created: broom", false, nil},
		{"\\list", "\nCurrent Members:\n\tLando Calrissian\n", false, nil},
		{"\\list-rooms", "\nCurrent rooms: \n  Room: broom\n  Members:\n\tLando Calrissian\n", false, nil},
		{"\\whoami", "\nClient Name: Lando Calrissian\nCurrent Room: broom", false, nil},
		{"\\leave", "You have left room broom", false, nil},
		{"\\invalid-command", "Invalid command: `\\invalid-command`", false, nil},
		{"\\exit", "Lando Calrissian has gone offline", true, io.EOF},
	}

	for _, tt := range tests {
		actualStr, actualBool, actualError := c.parseResponse(tt.input)
		assert.Equal(t, tt.expectedStr, actualStr)
		assert.Equal(t, tt.expectedBool, actualBool)
		assert.Equal(t, tt.expectedError, actualError)
	}
}

func Test_parseResponse_more_than_one_client_required(t *testing.T) {
	cm := &mocks.CacheMock{}
	c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: &mocks.IoWriterMock{}, Cache: cm}
	c2 := &Client{Id: "123", Name: "Leia Organa", CurrentRoom: "vroom", Writer: &mocks.IoWriterMock{}, Cache: cm}
	cm.GetMock = func(k string) (interface{}, bool) {
		return map[string][]*Client{"broom": {c1}, "vroom": {c2}}, true
	}

	actualStr1, actualBool1, actualError1 := c1.parseResponse("\\list vroom")
	actualStr2, actualBool2, actualError2 := c2.parseResponse("\\list broom")

	actualStr3, actualBool3, actualError3 := c2.parseResponse("\\join broom")

	assert.Equal(t, "\nCurrent Members:\n\tLeia Organa\n", actualStr1)
	assert.False(t, actualBool1)
	assert.Nil(t, actualError1)
	assert.Equal(t, "\nCurrent Members:\n\tHan Solo\n", actualStr2)
	assert.False(t, actualBool2)
	assert.Nil(t, actualError2)

	assert.Equal(t, "Leia Organa has entered: broom", actualStr3)
	assert.True(t, actualBool3)
	assert.Nil(t, actualError3)
}

func Test_listen_msg_broadcasts_to_room(t *testing.T) {
	count := 0
	monkey.Patch(Read, func(a interfaces.AbstractBufioReader) (string, error) {
		if count == 0 {
			count++
			return "test", nil
		} else {
			return "", io.EOF //Also secretly testing that io.EOF kills the process
		}
	})
	defer monkey.Unpatch(Read)

	monkey.Patch(time.Now, func() time.Time {
		return time.Date(2022, 04, 20, 11, 00, 00, 00, time.UTC)
	})
	defer monkey.Unpatch(time.Now)

	wg1 := sync.WaitGroup{}
	wg2 := sync.WaitGroup{}
	m1 := &mocks.IoWriterMock{WriteMock: func(p []byte) (n int, err error) {
		wg1.Done()
		return 0, nil
	}}
	m2 := &mocks.IoWriterMock{WriteMock: func(p []byte) (n int, err error) {
		wg2.Done()
		return 0, nil
	}}

	cm := &mocks.CacheMock{}

	c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: m1, Conn: &mocks.NetConnMock{}, Cache: cm}
	c2 := &Client{Id: "456", Name: "Leia Organa", CurrentRoom: "broom", Writer: m2, Conn: &mocks.NetConnMock{}, Cache: cm}
	cm.GetMock = func(k string) (interface{}, bool) {
		if k == CLIENTS {
			return map[string]*Client{c1.Id: c1, c2.Id: c2}, true
		} else {
			return map[string][]*Client{"broom": {c1, c2}}, true
		}
	}

	wg1.Add(1)
	wg2.Add(1)

	c1.listen()

	wg1.Wait()
	wg2.Wait()

	assert.True(t, m1.WriteCalled)
	assert.True(t, m2.WriteCalled)
	assert.Equal(t, "1650452400: Han Solo> test\n", string(m1.WriteCalledWith))
	assert.Equal(t, "1650452400: Han Solo: test\n", string(m2.WriteCalledWith))
}

func Test_listen_msg_sends_only_to_self(t *testing.T) {
	count := 0
	monkey.Patch(Read, func(a interfaces.AbstractBufioReader) (string, error) {
		if count == 0 {
			count++
			return "test", nil
		} else {
			return "", io.EOF
		}
	})
	defer monkey.Unpatch(Read)

	monkey.Patch(time.Now, func() time.Time {
		return time.Date(2022, 04, 20, 11, 00, 00, 00, time.UTC)
	})
	defer monkey.Unpatch(time.Now)

	m1 := &mocks.IoWriterMock{}
	m2 := &mocks.IoWriterMock{}
	cm := &mocks.CacheMock{}

	c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "", Writer: m1, Conn: &mocks.NetConnMock{}, Cache: cm}
	c2 := &Client{Id: "456", Name: "Leia Organa", CurrentRoom: "broom", Writer: m2, Conn: &mocks.NetConnMock{}, Cache: cm}
	cm.GetMock = func(k string) (interface{}, bool) {
		if k == CLIENTS {
			return map[string]*Client{c1.Id: c1, c2.Id: c2}, true
		} else {
			return map[string][]*Client{"broom": {c1, c2}}, true
		}
	}
	c1.listen()

	assert.True(t, m1.WriteCalled)
	assert.False(t, m2.WriteCalled)
	assert.Equal(t, "1650452400: Han Solo> test\n", string(m1.WriteCalledWith))
}

func Test_listen_cmd_broadcasts_to_room(t *testing.T) {
	count := 0
	monkey.Patch(Read, func(a interfaces.AbstractBufioReader) (string, error) {
		if count == 0 {
			count++
			return "\\name LukeSkywalker", nil
		} else {
			return "", io.EOF //Also secretly testing that io.EOF kills the process
		}
	})
	defer monkey.Unpatch(Read)

	monkey.Patch(time.Now, func() time.Time {
		return time.Date(2022, 04, 20, 11, 00, 00, 00, time.UTC)
	})
	defer monkey.Unpatch(time.Now)

	wg1 := sync.WaitGroup{}
	wg2 := sync.WaitGroup{}
	m1 := &mocks.IoWriterMock{WriteMock: func(p []byte) (n int, err error) {
		wg1.Done()
		return 0, nil
	}}
	m2 := &mocks.IoWriterMock{WriteMock: func(p []byte) (n int, err error) {
		wg2.Done()
		return 0, nil
	}}
	cm := &mocks.CacheMock{}

	c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: m1, Conn: &mocks.NetConnMock{}, Cache: cm}
	c2 := &Client{Id: "456", Name: "Leia Organa", CurrentRoom: "broom", Writer: m2, Conn: &mocks.NetConnMock{}, Cache: cm}
	cm.GetMock = func(k string) (interface{}, bool) {
		if k == CLIENTS {
			return map[string]*Client{c1.Id: c1, c2.Id: c2}, true
		} else {
			return map[string][]*Client{"broom": {c1, c2}}, true
		}
	}

	wg1.Add(1)
	wg2.Add(1)

	c1.listen()

	wg1.Wait()
	wg2.Wait()

	assert.True(t, m1.WriteCalled)
	assert.True(t, m2.WriteCalled)
	assert.Equal(t, "1650452400: LukeSkywalker> User: Han Solo has become -> LukeSkywalker\n", string(m1.WriteCalledWith))
	assert.Equal(t, "1650452400: LukeSkywalker: User: Han Solo has become -> LukeSkywalker\n", string(m2.WriteCalledWith))
}

func Test_listen_cmd_sends_only_to_self(t *testing.T) {
	count := 0
	monkey.Patch(Read, func(a interfaces.AbstractBufioReader) (string, error) {
		if count == 0 {
			count++
			return "\\list", nil
		} else {
			return "", io.EOF
		}
	})
	defer monkey.Unpatch(Read)
	monkey.Patch(time.Now, func() time.Time {
		return time.Date(2022, 04, 20, 11, 00, 00, 00, time.UTC)
	})
	defer monkey.Unpatch(time.Now)

	m1 := &mocks.IoWriterMock{}
	m2 := &mocks.IoWriterMock{}
	cm := &mocks.CacheMock{}

	c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: m1, Conn: &mocks.NetConnMock{}, Cache: cm}
	c2 := &Client{Id: "456", Name: "Leia Organa", CurrentRoom: "broom", Writer: m2, Conn: &mocks.NetConnMock{}, Cache: cm}
	cm.GetMock = func(k string) (interface{}, bool) {
		if k == CLIENTS {
			return map[string]*Client{c1.Id: c1, c2.Id: c2}, true
		} else {
			return map[string][]*Client{"broom": {c1, c2}}, true
		}
	}
	c1.listen()

	assert.True(t, m1.WriteCalled)
	assert.False(t, m2.WriteCalled)
	assert.Equal(t, "1650452400: Han Solo> \nCurrent Members:\n\tHan Solo\n\tLeia Organa\n\n", string(m1.WriteCalledWith))
}
