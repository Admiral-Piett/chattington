package mocks

import (
    "bufio"
    "github.com/Admiral-Piett/chat-telnet/interfaces"
    "net"
    "time"
)

// ---- Server Mocks ----

type ServerMock struct {
    Listener net.Listener
    CloseCalled bool
    StartCalled bool
    CreateClientCalled bool
    ListenCalled bool
    CloseMock func()
    StartMock func()
}

func (m *ServerMock) Close() {
    m.CloseCalled = true
    if m.CloseMock != nil {
        m.CloseMock()
    }
}

func (m *ServerMock) Start() {
    m.StartCalled = true
    if m.StartMock != nil {
        m.StartMock()
    }
}

func (m *ServerMock) createClient(conn net.Conn) {
    m.CreateClientCalled = true
}

func (m *ServerMock) listen(conn net.Conn) {
    m.ListenCalled = true
}

type NetAddrMock struct {
    NetworkMock func() string
    StringMock func() string
}

func (m *NetAddrMock) Network() string {
    return ""
}
func (m *NetAddrMock) String() string {
    return ""
}

type NetListenerMock struct {
    AcceptCalled bool
    AcceptMock func() (net.Conn, error)
    CloseMock func() error
    AddrMock func() net.Addr
}

func (m *NetListenerMock) Accept() (net.Conn, error) {
    m.AcceptCalled = true
    if m.AcceptMock != nil {
        return m.AcceptMock()
    }
    return &NetConnMock{}, nil
}

func (m *NetListenerMock) Close() error {
    if m.CloseMock != nil {
        return m.CloseMock()
    }
    return nil
}

func (m *NetListenerMock) Addr() net.Addr {
    if m.AddrMock != nil {
        return m.AddrMock()
    }
    return &NetAddrMock{}
}

// ---- Client Mocks ----

type ClientMock struct {
    WriteStringCalled bool
    WriteResponseCalled bool
    Conn   interfaces.AbstractNetConn
    Name   string
    CurrentRoom string
    WriteStringMock func(msg string) error
    WriteResponseMock func(msg string, sendingClient interface{}) error
}

func (m *ClientMock) WriteString(msg string) error {
    m.WriteStringCalled = true
    if m.WriteStringMock != nil {
        return m.WriteStringMock(msg)
    }
    return nil
}

func (m *ClientMock) WriteResponse(msg string, sendingClient interface{}) error {
    m.WriteResponseCalled = true
    if m.WriteResponseMock != nil {
        return m.WriteResponseMock(msg, sendingClient)
    }
    return nil
}

type NetConnMock struct {
    WriteCalled bool
    CalledWith []byte
    ReadMock func(b []byte) (n int, err error)
    WriteMock func(b []byte) (n int, err error)
    RemoteAddrMock func() net.Addr
    CloseMock func() error
    LocalAddrMock func() net.Addr
    SetDeadlineMock func(t time.Time) error
    SetReadDeadlineMock func(t time.Time) error
    SetWriteDeadlineMock func(t time.Time) error
}

func (m *NetConnMock) Write(p []byte) (n int, err error) {
    m.WriteCalled = true
    m.CalledWith = p
    if m.WriteMock != nil {
        return m.WriteMock(p)
    }
    return 0, nil
}

func (m *NetConnMock) Read(p []byte) (n int, err error) {
    if m.WriteMock != nil {
        return m.WriteMock(p)
    }
    return 0, nil
}

func (m *NetConnMock) RemoteAddr() net.Addr {
    return &NetAddrMock{}
}

func (m *NetConnMock) Close() error {
    return nil
}

func (m *NetConnMock) LocalAddr() net.Addr {
    return &NetAddrMock{}
}
func (m *NetConnMock) SetDeadline(t time.Time) error {
    return nil
}
func (m *NetConnMock) SetReadDeadline(t time.Time) error {
    return nil
}
func (m *NetConnMock) SetWriteDeadline(t time.Time) error {
    return nil
}

type ReaderMock struct {
    ReadStringCalled bool
    bufio.Reader
    ReadMock func(delim byte) (string, error)
}

func (m *ReaderMock)ReadString(delim byte) (string, error) {
    m.ReadStringCalled = true
    if m.ReadMock != nil {
        return m.ReadMock(delim)
    }
    return "", nil
}