package client

import (
    "bufio"
    "fmt"
    "io"
    "net"
    "strings"
    "time"
)

type Client struct {
    writer io.Writer
    Conn   net.Conn
    Name   string
    CurrentRoom string
}

func NewClient(conn net.Conn) (*Client, string) {
    // Generate a semi-random id for ourselves, using a `---` pattern to start.  We can lean on this for now, to
    //identify users who haven't yet given their name but still access the clients by key.
    intialName := fmt.Sprintf("%v", time.Now().Unix())

    return &Client{
        Conn: conn,
        writer: conn,
        Name: intialName,
        CurrentRoom: "",
    }, intialName
}

func (c *Client) WriteString(msg string) error {
    _, err := c.writer.Write([]byte(msg))

    return err
}

func (c *Client) WriteResponse(msg string, sendingClient interface{}) error {
    // Using `sendingClient` you can send in the string name who the sender of the message is, if it's `nil` we'll
    //  or the same as the target client (`c`) then we can format the response as i.
    prefix := ""
    if sendingClient == nil || sendingClient == c.Name {
        prefix = fmt.Sprintf("%s>", c.Name)
    } else {
        prefix = fmt.Sprintf("%s:", sendingClient)
    }
    // Add chat room response formatting
    msg = fmt.Sprintf("%s %s\n", prefix, msg)
    return c.WriteString(msg)
}

type ChatReader struct {

}

// Should I attach this to a struct?
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
