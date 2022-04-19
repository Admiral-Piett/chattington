package clients

import (
   "bou.ke/monkey"
   "bufio"
   "fmt"
   "github.com/Admiral-Piett/chat-telnet/interfaces"
   "github.com/Admiral-Piett/chat-telnet/mocks"
   "github.com/stretchr/testify/assert"
   "io"
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
   w2 := &mocks.IoWriterMock{WithWaitGroup: true}
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
   w1 := &mocks.IoWriterMock{WithWaitGroup: true}
   w2 := &mocks.IoWriterMock{WithWaitGroup: true}
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
  w1 := &mocks.IoWriterMock{WithWaitGroup: true}
  c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: w1}
  ChatCache.Rooms["broom"] = []*Client{c1}

  w1.Wg.Add(1)

  c1.broadcastToRoom("test", "broom")

  w1.Wg.Wait()

  assert.Equal(t, "Han Solo> test\n", string(w1.WriteCalledWith))
}

func Test_parseResponse_one_client_required(t *testing.T) {
   var tests = []struct {
     input string
     expectedStr string
     expectedBool bool
   }{
     {"\\name Lando Calrissian", "User: Han Solo has become -> Lando Calrissian", true},
     {"\\create broom", "New room created: broom", false},
     {"\\list", "\nCurrent Members:\n\tLando Calrissian\n", false},
     {"\\list-rooms", "\nCurrent rooms: \n  Room: broom\n  Members:\n\tLando Calrissian\n", false},
     {"\\whoami", "\nClient Name: Lando Calrissian\nCurrent Room: broom", false},
     {"\\leave", "You have left room broom", false},
     {"\\invalid-command", "Invalid command: `\\invalid-command`", false},
   }
   c := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "", Writer: &mocks.IoWriterMock{}}

   for _, tt := range tests {
      actualStr, actualBool := c.parseResponse(tt.input)
      assert.Equal(t, tt.expectedStr, actualStr)
      assert.Equal(t, tt.expectedBool, actualBool)
   }
}

func Test_parseResponse_more_than_one_client_required(t *testing.T) {
  c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: &mocks.IoWriterMock{}}
  c2 := &Client{Id: "123", Name: "Leia Organa", CurrentRoom: "vroom", Writer: &mocks.IoWriterMock{}}

  ChatCache.Rooms["broom"] = []*Client{c1}
  ChatCache.Rooms["vroom"] = []*Client{c2}

  actualStr1, actualBool1 := c1.parseResponse("\\list vroom")
  actualStr2, actualBool2 := c2.parseResponse("\\list broom")

  actualStr3, actualBool3 := c2.parseResponse("\\join broom")

  assert.Equal(t, "\nCurrent Members:\n\tLeia Organa\n", actualStr1)
  assert.False(t, actualBool1)
  assert.Equal(t, "\nCurrent Members:\n\tHan Solo\n", actualStr2)
  assert.False(t, actualBool2)

  assert.Equal(t, "Leia Organa has entered: broom", actualStr3)
  assert.True(t, actualBool3)
}

func Test_listen_msg_broadcasts_to_room(t *testing.T) {
  count := 0
  monkey.Patch(Read, func(a interfaces.AbstractBufioReader) (string, error) {
     if count == 0 {
        count ++
        return "test", nil
     } else {
        return "", io.EOF  //Also secretly testing that io.EOF kills the process
     }
  })
  defer monkey.Unpatch(bufio.NewReader)

  m1 := &mocks.IoWriterMock{}
  m2 := &mocks.IoWriterMock{}

  c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: m1, Conn: &mocks.NetConnMock{}}
  c2 := &Client{Id: "456", Name: "Leia Organa", CurrentRoom: "broom", Writer: m2, Conn: &mocks.NetConnMock{}}
  ChatCache.Clients[c1.Id] = c1
  ChatCache.Clients[c2.Id] = c2
  ChatCache.Rooms["broom"] = []*Client{c1, c2}
  c1.listen()

  assert.True(t, m1.WriteCalled)
  assert.True(t, m2.WriteCalled)
  assert.Equal(t, "Han Solo> test\n", string(m1.WriteCalledWith))
  assert.Equal(t, "Han Solo: test\n", string(m2.WriteCalledWith))
}

func Test_listen_msg_sends_only_to_self(t *testing.T) {
   count := 0
   monkey.Patch(Read, func(a interfaces.AbstractBufioReader) (string, error) {
      if count == 0 {
         count ++
         return "test", nil
      } else {
         return "", io.EOF
      }
   })
   defer monkey.Unpatch(bufio.NewReader)

   m1 := &mocks.IoWriterMock{}
   m2 := &mocks.IoWriterMock{}

   c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "", Writer: m1, Conn: &mocks.NetConnMock{}}
   c2 := &Client{Id: "456", Name: "Leia Organa", CurrentRoom: "broom", Writer: m2, Conn: &mocks.NetConnMock{}}
   ChatCache.Clients[c1.Id] = c1
   ChatCache.Clients[c2.Id] = c2
   ChatCache.Rooms["broom"] = []*Client{c1, c2}
   c1.listen()

   assert.True(t, m1.WriteCalled)
   assert.False(t, m2.WriteCalled)
   assert.Equal(t, "Han Solo> test\n", string(m1.WriteCalledWith))
}

func Test_listen_cmd_broadcasts_to_room(t *testing.T) {
   count := 0
   monkey.Patch(Read, func(a interfaces.AbstractBufioReader) (string, error) {
      if count == 0 {
         count ++
         return "\\name LukeSkywalker", nil
      } else {
         return "", io.EOF  //Also secretly testing that io.EOF kills the process
      }
   })
   defer monkey.Unpatch(bufio.NewReader)

   m1 := &mocks.IoWriterMock{}
   m2 := &mocks.IoWriterMock{}

   c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: m1, Conn: &mocks.NetConnMock{}}
   c2 := &Client{Id: "456", Name: "Leia Organa", CurrentRoom: "broom", Writer: m2, Conn: &mocks.NetConnMock{}}
   ChatCache.Clients[c1.Id] = c1
   ChatCache.Clients[c2.Id] = c2
   ChatCache.Rooms["broom"] = []*Client{c1, c2}
   c1.listen()

   assert.True(t, m1.WriteCalled)
   assert.True(t, m2.WriteCalled)
   assert.Equal(t, "LukeSkywalker> User: Han Solo has become -> LukeSkywalker\n", string(m1.WriteCalledWith))
   assert.Equal(t, "LukeSkywalker: User: Han Solo has become -> LukeSkywalker\n", string(m2.WriteCalledWith))
}

func Test_listen_cmd_sends_only_to_self(t *testing.T) {
   count := 0
   monkey.Patch(Read, func(a interfaces.AbstractBufioReader) (string, error) {
      if count == 0 {
         count ++
         return "\\list", nil
      } else {
         return "", io.EOF
      }
   })
   defer monkey.Unpatch(bufio.NewReader)

   m1 := &mocks.IoWriterMock{}
   m2 := &mocks.IoWriterMock{}

   c1 := &Client{Id: "123", Name: "Han Solo", CurrentRoom: "broom", Writer: m1, Conn: &mocks.NetConnMock{}}
   c2 := &Client{Id: "456", Name: "Leia Organa", CurrentRoom: "broom", Writer: m2, Conn: &mocks.NetConnMock{}}
   ChatCache.Clients[c1.Id] = c1
   ChatCache.Clients[c2.Id] = c2
   ChatCache.Rooms["broom"] = []*Client{c1, c2}
   c1.listen()

   assert.True(t, m1.WriteCalled)
   assert.False(t, m2.WriteCalled)
   assert.Equal(t, "Han Solo> \nCurrent Members:\n\tHan Solo\n\tLeia Organa\n\n", string(m1.WriteCalledWith))
}
