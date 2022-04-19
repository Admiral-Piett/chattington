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
	"time"
)

func Test_NewServer_success(t *testing.T) {
	monkey.Patch(net.Listen, func(a, b string) (net.Listener, error) {
		return &mocks.NetListenerMock{}, nil
	})
	defer monkey.Unpatch(net.Listen)
	s, err := servers.NewServer()

	assert.Nil(t, err)
	assert.IsType(t, servers.Server{}, s)
}

func Test_NewServer_net_listen_error(t *testing.T) {
	monkey.Patch(net.Listen, func(a, b string) (net.Listener, error) {
		return &mocks.NetListenerMock{}, fmt.Errorf("boom")
	})
	defer monkey.Unpatch(net.Listen)
	_, err := servers.NewServer()

	assert.Error(t, err)
}

func Test_Close_success(t *testing.T) {
	l := &mocks.NetListenerMock{}
	m := servers.Server{
		Listener: l,
	}
	m.Close()

	assert.True(t, l.CloseCalled)
}

func Test_Start_success(t *testing.T) {
	l := &mocks.NetListenerMock{}
	m := servers.Server{
		Listener: l,
	}

	patchCalled := false
	monkey.Patch(clients.GenerateNewClient, func(conn interfaces.AbstractNetConn) {
		patchCalled = true
	})
	defer monkey.Unpatch(net.Listen)

	go m.Start()
	time.Sleep(1 * time.Second) //TODO - find something more elegant than this

	assert.True(t, l.AcceptCalled)
	assert.True(t, patchCalled)
}

func Test_Start_Accept_raises_error(t *testing.T) {
	l := &mocks.NetListenerMock{}
	l.AcceptMock = func() (net.Conn, error) {
		return &mocks.NetConnMock{}, fmt.Errorf("boom")
	}
	m := servers.Server{
		Listener: l,
	}
	err := m.Start()

	assert.Error(t, err)
}
