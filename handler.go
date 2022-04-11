package main

import (
    "github.com/reiver/go-oi"
    t "github.com/reiver/go-telnet"
)

type ChatHandler struct{}

var clients = make(map[string]string)

func (h ChatHandler) ServeTELNET(ctx t.Context, w t.Writer, r t.Reader) {
    var buffer [1]byte // Seems like the length of the buffer needs to be small, otherwise will have to wait for buffer to fill up.
    p := buffer[:]

    for {
        n, err := r.Read(p)

        if n > 0 {
            oi.LongWrite(w, p[:n])
        }

        if nil != err {
            break
        }
    }
}

