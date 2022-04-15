package mocks

import (
    "bufio"
    "net"
    "time"
)

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
    CloseCalled bool
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
    m.CloseCalled = true
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

type IoWriterMock struct {
    WriteCalledWith []byte
    WriteMock       func(p []byte) (n int, err error)
}

func (m *IoWriterMock)Write(p []byte) (n int, err error) {
    m.WriteCalledWith = p
    if m.WriteMock != nil {
        return m.WriteMock(p)
    }
    return 0, nil
}