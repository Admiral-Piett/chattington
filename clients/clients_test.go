package clients_test

import (
   "fmt"
   "github.com/Admiral-Piett/chat-telnet/clients"
   "github.com/Admiral-Piett/chat-telnet/mocks"
   "github.com/stretchr/testify/assert"
   "testing"
)

func Test_WriteString_success(t *testing.T) {
   w := &mocks.IoWriterMock{}
   m := &clients.Client{
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
   m := &clients.Client{
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
   m := &clients.Client{
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
   m := &clients.Client{
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
   m := &clients.Client{
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
  _, err := clients.Read(m)

  assert.Nil(t, err)
  assert.True(t, m.ReadStringCalled)
}

func Test_Read_returns_error(t *testing.T) {
  m := &mocks.ReaderMock{
      ReadMock: func(delim byte) (string, error) {
          return "", fmt.Errorf("boom")
      },
  }
  s, err := clients.Read(m)

  assert.Error(t, err)
  assert.Empty(t, s)
  assert.True(t, m.ReadStringCalled)
}
