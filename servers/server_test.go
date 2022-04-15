package servers_test

import (
    "bou.ke/monkey"
    "fmt"
    "github.com/Admiral-Piett/chat-telnet/clients"
    "github.com/Admiral-Piett/chat-telnet/interfaces"
    "github.com/Admiral-Piett/chat-telnet/mocks"
    "github.com/Admiral-Piett/chat-telnet/servers"
    "github.com/stretchr/testify/assert"
    "net"
    "testing"
)

func Test_NewServer_success(t *testing.T) {
    monkey.Patch(net.Listen, func(a, b string) (net.Listener, error) {
        return &mocks.NetListerMock{}, nil
    })
    defer monkey.Unpatch(net.Listen)
    s, err := servers.NewServer()

    assert.Nil(t, err)
    assert.IsType(t, servers.Server{}, s)
}

func Test_NewServer_net_listen_error(t *testing.T) {
    monkey.Patch(net.Listen, func(a, b string) (net.Listener, error) {
        return &mocks.NetListerMock{}, fmt.Errorf("boom")
    })
    defer monkey.Unpatch(net.Listen)
    _, err := servers.NewServer()

    assert.Nil(t, err)
}

func Test_Close_success(t *testing.T) {
    m := mocks.ServerMock{}
    m.Close()

    assert.True(t, m.CloseCalled)
}

func Test_Start_success(t *testing.T) {
    l := &mocks.NetListerMock{}
    m, _ := servers.NewServer()
    m.Listener = l

    c := &mocks.ClientMock{
        Conn: &mocks.NetConnMock{},
        Name: "test",
    }

    m.Start()

    assert.True(t, l.AcceptCalled)
}

func Test_Start_Accept_raises_error(t *testing.T) {
    m := mocks.ServerMock{}
    m.Close()

    assert.True(t, m.CloseCalled)
}


