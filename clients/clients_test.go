package clients_test
//
//import (
//    "fmt"
//    "github.com/Admiral-Piett/chat-telnet/clients"
//    "github.com/Admiral-Piett/chat-telnet/mocks"
//    "github.com/stretchr/testify/assert"
//    "reflect"
//    "strings"
//    "testing"
//)
//
//
//func Test_WriteString_success(t *testing.T) {
//    expectedClient := clients.Client{}
//
//    m := &mocks.NetConnMock{}
//    c, name := clients.GenerateNewClient(m)
//
//    if reflect.TypeOf(*c) != reflect.TypeOf(expectedClient) {
//        t.Errorf("Expected clients.Client{} to have been generated")
//    }
//
//    err := c.WriteString("test")
//
//    assert.Nil(t, err)
//    assert.True(t, m.WriteCalled)
//    assert.IsType(t, "", name)
//}
//
//func Test_WriteString_returns_error(t *testing.T) {
//    m := &mocks.NetConnMock{
//        WriteMock: func(b []byte) (n int, err error) {
//            return 0, fmt.Errorf("boom")
//        },
//    }
//    c, _ := clients.GenerateNewClient(m)
//    err := c.WriteString("test")
//
//    assert.Error(t, err)
//}
//
//func Test_WriteResponse_success_no_sendingClient(t *testing.T) {
//    m := &mocks.NetConnMock{}
//    c, _ := clients.GenerateNewClient(m)
//
//    err := c.WriteResponse("test", nil)
//
//    assert.Nil(t, err)
//    if !strings.HasSuffix(string(m.CalledWith), "> test\n") {
//       t.Errorf("Expected Write() to have ended with `> test\\n`, NOT: %s", string(m.CalledWith))
//    }
//}
//
//func Test_WriteResponse_success_with_sendingClient(t *testing.T) {
//    m := &mocks.NetConnMock{}
//    c, _ := clients.GenerateNewClient(m)
//
//    err := c.WriteResponse("test", "HanSolo")
//
//    assert.Nil(t, err)
//    assert.Equal(t, "HanSolo: test\n", m.CalledWith)
//}
//
//func Test_WriteResponse_returns_error(t *testing.T) {
//    m := &mocks.NetConnMock{
//        WriteMock: func(b []byte) (n int, err error) {
//            return 0, fmt.Errorf("boom")
//        },
//    }
//    c, _ := clients.GenerateNewClient(m)
//
//    err := c.WriteResponse("test", "HanSolo")
//
//    assert.Error(t, err)
//}
//
//func Test_Read_success(t *testing.T) {
//    m := &mocks.ReaderMock{}
//    _, err := clients.Read(m)
//
//    assert.Error(t, err)
//    assert.True(t, m.ReadStringCalled)
//}
//
//func Test_Read_returns_error(t *testing.T) {
//    m := &mocks.ReaderMock{
//        ReadMock: func(delim byte) (string, error) {
//            return "", fmt.Errorf("boom")
//        },
//    }
//    s, err := clients.Read(m)
//
//    assert.Error(t, err)
//    assert.Empty(t, s)
//    assert.True(t, m.ReadStringCalled)
//}
