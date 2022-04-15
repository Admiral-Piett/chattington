package main

import (
    "github.com/Admiral-Piett/chat-telnet/servers"
    "log"
)

func main() {
    s, err := servers.NewServer()
    if err != nil {
        log.Fatal(err)
    }
    defer s.Close()

    err = s.Start()
    if err != nil {
        log.Fatal(err)
    }
}
