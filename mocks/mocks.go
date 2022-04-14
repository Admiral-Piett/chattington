package mocks

import (
    "bufio"
    "net"
)

type NetConnMock struct {
    WriteCalled bool
    CalledWith []byte
    ReadMock func(b []byte) (n int, err error)
    WriteMock func(b []byte) (n int, err error)
    RemoteAddrMock func() net.Addr
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