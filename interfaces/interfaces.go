package interfaces

import (
	"io"
	"net"
	"time"
)

// bufio.Reader doesn't implement it's own interface so we will do that for testing really.
type AbstractBufioReader interface {
	io.Reader
	ReadString(delim byte) (string, error)
}

type AbstractClient interface {
	WriteString(msg string) error
	WriteResponse(msg string, sendingClient interface{}) error
}

type AbstractCache interface {
	Get(k string) (interface{}, bool)
	Set(k string, x interface{}, d time.Duration)
}

type AbstractIoWriter interface {
	Write(p []byte) (n int, err error)
}

type AbstractNetConn interface {
	Close() error
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	RemoteAddr() net.Addr
}

type AbstractServer interface {
	Close()
	Start()
}
