package main

import (
    "bufio"
    "errors"
    "io"
    "log"
    "strings"
)

type ChatWriter struct {
    writer io.Writer
}

func NewCommandWriter(writer io.Writer) *ChatWriter {
    return &ChatWriter{
        writer: writer,
    }
}

func (w *ChatWriter) WriteString(msg string) error {
    _, err := w.writer.Write([]byte(msg))

    return err
}

type ChatReader struct {
    reader *bufio.Reader
}

func NewCommandReader(reader io.Reader) *ChatReader {
    return &ChatReader{
        reader: bufio.NewReader(reader),
    }
}

func (r *ChatReader) Read() (string, error) {
    value, err := r.reader.ReadString('\n')
    // Looks like can usually expect an io.EOF on connection death here, or potential bad characters, etc.  Either
    //way, log it back to the user, so they can see it and stop putting garbage back (assuming they haven't
    //killed their connection in which case it won't matter).
    if err != nil {
        return "", err
    }

    value = strings.TrimSpace(value)

    cmd := strings.Split(value, " ")[0]
    // TODO - add other commands in here
    switch {
    // FIXME - this is a little ham-fisted.  Come up with another way of not blocking a pattern of user input.
    case strings.HasPrefix(cmd, "---"):
        log.Printf("Invalid command: %s\n", value)
        return "", errors.New("Invalid input")
    case cmd == "\\name":
        log.Printf("Name Command: %s\n", value)
    default:
        log.Printf("Unknown Command: `%s`\n", value)
    }

    return value, err
}
