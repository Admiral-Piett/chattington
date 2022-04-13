package main

import (
    "bufio"
    "fmt"
    "io"
    "net"
    "strings"
)

type Client struct {
    conn   net.Conn
    name   string
    writer io.Writer
    CurrentRoom string
}

func (c *Client) WriteString(msg string) error {
    _, err := c.writer.Write([]byte(msg))

    return err
}

func (c *Client) WriteResponse(msg string, targetClient interface{}) error {
    // Using `targetClient` you can send in the string name who the sender of the message is, if it's `nil` we'll
    //just assume it's the same as the client that is receiving the message.
    if targetClient == nil {
        targetClient = c.name
    }
    // Add chat room response formatting
    msg = fmt.Sprintf("%s: %s\n%s> ", targetClient, msg, c.name)
    return c.WriteString(msg)
}

// Should I attach this to a struct
func Read(r *bufio.Reader) (string, error) {
    value, err := r.ReadString('\n')
    // Looks like can usually expect an io.EOF on connection death here, or potential bad characters, etc.  Either
    //way, log it back to the user, so they can see it and stop putting garbage back (assuming they haven't
    //killed their connection in which case it won't matter).
    if err != nil {
        return "", err
    }

    value = strings.TrimSpace(value)
    return value, err
}
