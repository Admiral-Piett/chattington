package main

import (
    "log"
)

func main() {
    s, err := NewServer()
    if err != nil {
        log.Fatal(err)
    }
    defer s.Close()
    s.rooms["fart"] = []string{"blah", "foo"}
    s.Start()
}
