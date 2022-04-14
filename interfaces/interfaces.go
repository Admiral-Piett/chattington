package interfaces

import (
    "io"
    "net"
)

type AbstractClient interface {
    WriteString(msg string) error
    WriteResponse(msg string, sendingClient interface{}) error
}

type AbstractIoWriter interface {
    Write(p []byte) (n int, err error)
}

// Only have to implement what we need
type AbstractNetConn interface {
    Read(b []byte) (n int, err error)
    Write(b []byte) (n int, err error)
    RemoteAddr() net.Addr
}


// bufio.Reader doesn't implement it's own interface so we will do that for testing really.
type AbstractBufioReader interface {
    io.Reader
    ReadString(delim byte) (string, error)
}
