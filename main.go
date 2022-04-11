package main

import (
	"github.com/reiver/go-telnet"
)

func main() {
    handler := ChatHandler{}

    err := telnet.ListenAndServe(":23", handler)
    if nil != err {
        panic(err)
    }
}
