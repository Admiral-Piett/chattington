package servers_test

import (
	"bou.ke/monkey"
	"chat-telnet/clients"
	"chat-telnet/interfaces"
	"chat-telnet/mocks"
	"chat-telnet/servers"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net"
	"sync"
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
	wg := sync.WaitGroup{}
	wg.Add(1)

	patchCalled := false
	monkey.Patch(clients.GenerateNewClient, func(conn interfaces.AbstractNetConn, cache interfaces.AbstractCache) error {
		// This gets called in a loop that would, in real life hang, waiting for a connection.  So we'll hit the wg
		//a bunch of times before we finish waiting.  So just make sure we hit it at LEAST once, and simulate
		//the "hang" below.
		if !patchCalled {
			wg.Done()
		}
		patchCalled = true
		time.Sleep(1 * time.Second)
		return nil
	})
	defer monkey.Unpatch(net.Listen)

	go m.Start()
	wg.Wait()
	m.Close()

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

func Test_Start_GenerateNewClient_raises_error(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	patchCalled := false
	monkey.Patch(clients.GenerateNewClient, func(conn interfaces.AbstractNetConn, cache interfaces.AbstractCache) error {
		// This gets called in a loop that would, in real life hang, waiting for a connection.  So we'll hit the wg
		//a bunch of times before we finish waiting.  So just make sure we hit it at LEAST once, and simulate
		//the "hang" below.
		if !patchCalled {
			wg.Done()
		}
		patchCalled = true
		time.Sleep(1 * time.Second)
		return fmt.Errorf("boom")
	})
	defer monkey.Unpatch(net.Listen)

	l := &mocks.NetListenerMock{}
	m := servers.Server{
		Listener: l,
	}

	go m.Start()
	wg.Wait()
	m.Close()

	assert.True(t, patchCalled)
	assert.True(t, l.CloseCalled)
}
