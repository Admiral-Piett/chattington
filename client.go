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

func (c *Client) WriteResponse(msg string, sendingClient interface{}) error {
    // Using `sendingClient` you can send in the string name who the sender of the message is, if it's `nil` we'll
    //  or the same as the target client (`c`) then we can format the response as if it's from that person.
    prefix := ""
    if sendingClient == nil || sendingClient == c.name {
        prefix = fmt.Sprintf("%s>", c.name)
    } else {
        prefix = fmt.Sprintf("%s:", sendingClient)
    }
    // Add chat room response formatting
    msg = fmt.Sprintf("%s %s\n", prefix, msg)
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
