package clients

import (
   "fmt"
   "github.com/Admiral-Piett/chat-telnet/mocks"
   "github.com/stretchr/testify/assert"
   "os"
   "sync"
   "testing"
)

func TestMain(m *testing.M) {
   setup()
   code := m.Run()
   os.Exit(code)
}

func setup() {
   ChatCache = &ChatMeta{  // Best effort to reset the cache
     Clients:  map[string]*Client{},
     Rooms:    map[string][]*Client{},
     Mutex:    &sync.Mutex{},
   }
}

func Test_WriteString_success(t *testing.T) {
   w := &mocks.IoWriterMock{}
   m := &Client{
      Writer: w,
      Conn: &mocks.NetConnMock{},
      Name: "Han Solo",
      CurrentRoom: "",
      Id: "test-id",
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
      Writer: w,
      Conn: &mocks.NetConnMock{},
      Name: "Han Solo",
      CurrentRoom: "",
      Id: "test-id",
   }
   err := m.WriteString("Hi")

   assert.Error(t, err)
}

func Test_WriteResponse_success_no_sendingClient(t *testing.T) {
   w := &mocks.IoWriterMock{}
   m := &Client{
      Writer: w,
      Conn: &mocks.NetConnMock{},
      Name: "Han Solo",
      CurrentRoom: "",
      Id: "test-id",
   }
   err := m.WriteResponse("Hi", nil)

   assert.Nil(t, err)
   assert.Equal(t, "Han Solo> Hi\n", string(w.WriteCalledWith))
}

func Test_WriteResponse_success_with_sendingClient(t *testing.T) {
   w := &mocks.IoWriterMock{}
   m := &Client{
      Writer: w,
      Conn: &mocks.NetConnMock{},
      Name: "Han Solo",
      CurrentRoom: "",
      Id: "test-id",
   }
   err := m.WriteResponse("Hi", "Leia Organa")

   assert.Nil(t, err)
   assert.Equal(t, "Leia Organa: Hi\n", string(w.WriteCalledWith))
}

func Test_WriteResponse_returns_error(t *testing.T) {
   w := &mocks.IoWriterMock{
      WriteMock: func(p []byte) (n int, err error) {
         return 0, fmt.Errorf("boom")
      },
   }
   m := &Client{
      Writer: w,
      Conn: &mocks.NetConnMock{},
      Name: "Han Solo",
      CurrentRoom: "",
      Id: "test-id",
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
   c := &Client{Id: "123", Conn: &mocks.NetConnMock{}}
   ChatCache.Clients["123"] = c
   c.removeConnection()

   assert.NotContains(t, ChatCache.Clients, "123")
}

func Test_changeClientName_success(t *testing.T) {
   c := &Client{Id: "123", Name: "Han Solo"}
   ChatCache.Clients["123"] = c
   c.changeClientName("Luke Skywalker")

   assert.Equal(t, "Luke Skywalker", c.Name)
   assert.Equal(t, "123", c.Id)
}

func Test_displayClientStats_success(t *testing.T) {
   c := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom"}
   response, b := c.displayClientStats()

   assert.Equal(t, "\nClient Name: Han Solo\nCurrent Room: broom", response)
   assert.False(t, b)
}

func Test_listRooms_success(t *testing.T) {
   c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom"}
   c2 := &Client{Id: "456", Name: "Chewbacca", CurrentRoom: "broom"}
   ChatCache.Rooms["broom"] = []*Client{c1, c2}
   response, b := c1.listRooms()

   assert.Equal(t, "\nCurrent rooms: \n  Room: broom\n  Members:\n\tHan Solo\n\tChewbacca\n", response)
   assert.False(t, b)
}

func Test_listRooms_no_rooms(t *testing.T) {
   // FIXME - remove this after we wire in a real cache.  Otherwise, we have to force this "reset" right now,
   //  otherwise we can't guarantee that the Cache would be empty because the tests run concurrently.
   ChatCache = &ChatMeta{
     Clients:  map[string]*Client{},
     Rooms:    map[string][]*Client{},
     Mutex:    &sync.Mutex{},
   }
   c := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom"}
   response, b := c.listRooms()

   assert.Equal(t, "No rooms yet - make one!", response)
   assert.False(t, b)
}

func Test_listMembers_success(t *testing.T) {
   c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom"}
   c2 := &Client{Id: "456", Name: "Chewbacca", CurrentRoom: "broom"}
   ChatCache.Rooms["broom"] = []*Client{c1, c2}
   response, b := c1.listMembers("broom")

   assert.Equal(t, "\nCurrent Members:\n\tHan Solo\n\tChewbacca\n", response)
   assert.False(t, b)
}

func Test_listMembers_invalid_roomName(t *testing.T) {
   c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom"}
   c2 := &Client{Id: "456", Name: "Chewbacca", CurrentRoom: "broom"}
   ChatCache.Rooms["broom"] = []*Client{c1, c2}
   response, b := c1.listMembers("vroom")

   assert.Equal(t, "No such room vroom!", response)
   assert.False(t, b)
}

func Test_createRoom_success(t *testing.T) {
   c := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: &mocks.IoWriterMock{}}
   ChatCache.Rooms["broom"] = []*Client{c}

   response, b := c.createRoom("mushroom")

   assert.Equal(t, []*Client{c}, ChatCache.Rooms["mushroom"])
   assert.Equal(t, "mushroom", c.CurrentRoom)
   assert.NotContains(t, ChatCache.Rooms, "broom")
   assert.Equal(t, "New room created: mushroom", response)
   assert.False(t, b)
}

func Test_joinRoom_success(t *testing.T) {
   c := &Client{Id: "123", Name: "Han Solo", Writer: &mocks.IoWriterMock{}}
   ChatCache.Rooms["broom"] = []*Client{}

   response, b := c.joinRoom("broom")

   assert.Equal(t, []*Client{c}, ChatCache.Rooms["broom"])
   assert.Equal(t, "broom", c.CurrentRoom)
   assert.Equal(t, "Han Solo has entered: broom", response)
   assert.True(t, b)
}

func Test_joinRoom_room_does_not_exist(t *testing.T) {
   c := &Client{Id: "123", Name: "Han Solo", Writer: &mocks.IoWriterMock{}}
   ChatCache.Rooms["broom"] = []*Client{}

   response, b := c.joinRoom("vroom")

   assert.Equal(t, "", c.CurrentRoom)
   assert.Equal(t, "Room `vroom` doesn't exist - try creating it with `\\create`", response)
   assert.False(t, b)
}

func Test_joinRoom_already_in_room(t *testing.T) {
   c := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: &mocks.IoWriterMock{}}
   ChatCache.Rooms["broom"] = []*Client{c}

   response, b := c.joinRoom("broom")

   assert.Equal(t, "broom", c.CurrentRoom)
   assert.Equal(t, "You're already in broom!", response)
   assert.False(t, b)
}

func Test_leaveRoom_success(t *testing.T) {
   w2 := &mocks.IoWriterMock{}
   c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: &mocks.IoWriterMock{}}
   c2 := &Client{Id: "456", Name: "Chewbacca", CurrentRoom: "broom", Writer: w2}
   ChatCache.Rooms["broom"] = []*Client{c1, c2}

   w2.Wg.Add(1)

   c1.leaveRoom("broom")

   w2.Wg.Wait()

   assert.Equal(t, []*Client{c2}, ChatCache.Rooms["broom"])
   assert.Equal(t, "Han Solo: Han Solo has left broom.\n", string(w2.WriteCalledWith))
}

func Test_leaveRoom_empty_string_room_name(t *testing.T) {
   w2 := &mocks.IoWriterMock{}
   c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: &mocks.IoWriterMock{}}
   c2 := &Client{Id: "456", Name: "Chewbacca", CurrentRoom: "broom", Writer: w2}
   ChatCache.Rooms["broom"] = []*Client{c1, c2}

   c1.leaveRoom("broom")

   assert.False(t, w2.WriteCalled)
}

func Test_leaveRoom_empties_out_room_destroys_room(t *testing.T) {
   c := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: &mocks.IoWriterMock{}}
   ChatCache.Rooms["broom"] = []*Client{c}

   c.leaveRoom("broom")

   assert.Nil(t, ChatCache.Rooms["broom"])
}

func Test_broadcastToRoom_success(t *testing.T) {
   w1 := &mocks.IoWriterMock{}
   w2 := &mocks.IoWriterMock{}
   c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: w1}
   c2 := &Client{Id: "456", Name: "Chewbacca", CurrentRoom: "broom", Writer: w2}
   ChatCache.Rooms["broom"] = []*Client{c1, c2}

   w1.Wg.Add(1)
   w2.Wg.Add(1)

   c1.broadcastToRoom("test", "broom")

   w1.Wg.Wait()
   w2.Wg.Wait()

   assert.Equal(t, "Han Solo> test\n", string(w1.WriteCalledWith))
   assert.Equal(t, "Han Solo: test\n", string(w2.WriteCalledWith))
}

func Test_broadcastToRoom_alone_write_to_self(t *testing.T) {
   w1 := &mocks.IoWriterMock{}
   c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: w1}
   ChatCache.Rooms["broom"] = []*Client{c1}

   w1.Wg.Add(1)

   c1.broadcastToRoom("test", "broom")

   w1.Wg.Wait()

   assert.Equal(t, "Han Solo> test\n", string(w1.WriteCalledWith))
}

// TODO - HERE - Add listen and parseResponse tests
