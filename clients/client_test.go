package clients_test

import (
    "fmt"
    "github.com/Admiral-Piett/chat-telnet/clients"
    "github.com/Admiral-Piett/chat-telnet/mocks"
    "reflect"
    "strings"
    "testing"
)


func Test_WriteString_success(t *testing.T) {
    expectedClient := clients.Client{}

    m := &mocks.NetConnMock{}
    c, name := clients.NewClient(m)

    if reflect.TypeOf(*c) != reflect.TypeOf(expectedClient) {
        t.Errorf("Expected clients.Client{} to have been generated")
    }

    err := c.WriteString("test")

    if err != nil {
        t.Errorf("Unexpected error thrown: %s", err)
    }
    if m.WriteCalled != true {
       t.Errorf("Expected Write() to have been called")
    }
    if reflect.TypeOf(name) != reflect.TypeOf("") {
        t.Errorf("Expected name to be of type `int`")
    }
}

func Test_WriteString_returns_error(t *testing.T) {
    m := &mocks.NetConnMock{
        WriteMock: func(b []byte) (n int, err error) {
            return 0, fmt.Errorf("boom")
        },
    }
    c, _ := clients.NewClient(m)
    err := c.WriteString("test")

    if err == nil {
       t.Errorf("Expected error to be thrown")
    }
}

func Test_WriteResponse_success_no_sendingClient(t *testing.T) {
    m := &mocks.NetConnMock{}
    c, _ := clients.NewClient(m)

    err := c.WriteResponse("test", nil)

    if err != nil {
        t.Errorf("Unexpected error thrown: %s", err)
    }
    if !strings.HasSuffix(string(m.CalledWith), "> test\n") {
       t.Errorf("Expected Write() to have ended with `> test\\n`, NOT: %s", string(m.CalledWith))
    }
}

func Test_WriteResponse_success_with_sendingClient(t *testing.T) {
    m := &mocks.NetConnMock{}
    c, _ := clients.NewClient(m)

    err := c.WriteResponse("test", "HanSolo")

    if err != nil {
        t.Errorf("Unexpected error thrown: %s", err)
    }
    if string(m.CalledWith) != "HanSolo: test\n" {
        t.Errorf("Expected Write() to have ended with `: test\\n`, NOT: %s", string(m.CalledWith))
    }
}

func Test_WriteResponse_returns_error(t *testing.T) {
    m := &mocks.NetConnMock{
        WriteMock: func(b []byte) (n int, err error) {
            return 0, fmt.Errorf("boom")
        },
    }
    c, _ := clients.NewClient(m)

    err := c.WriteResponse("test", "HanSolo")

    if err == nil {
        t.Errorf("Expected error to be thrown")
    }
}

func Test_Read_success(t *testing.T) {
    m := &mocks.ReaderMock{}
    _, err := clients.Read(m)

    if err != nil {
        t.Errorf("Unexpected error thrown: %s", err)
    }
    if m.ReadStringCalled != true {
        t.Errorf("Expected ReadString() to be called")
    }
}

func Test_Read_returns_error(t *testing.T) {
    m := &mocks.ReaderMock{
        ReadMock: func(delim byte) (string, error) {
            return "", fmt.Errorf("boom")
        },
    }
    s, err := clients.Read(m)
    if m.ReadStringCalled != true {
        t.Errorf("Expected ReadString() to be called")
    }
    if s != "" || err == nil {
        t.Errorf("Expected error to be thrown")
    }
}
