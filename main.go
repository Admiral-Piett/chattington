package main

import (
    "github.com/Admiral-Piett/chat-telnet/server"
    "log"
)

func main() {
    s, err := server.NewServer()
    if err != nil {
        log.Fatal(err)
    }
    defer s.Close()

    s.Start()
}
