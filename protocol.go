package main

import (
    "bufio"
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
    if err != nil {
        return "", err
    }

    value = strings.TrimSpace(value)

    cmd := strings.Split(value, " ")[0]
    // TODO - add other commands in here
    switch {
    case cmd == "\\name":
        log.Printf("Name Command: %s\n", value)
    default:
        log.Printf("Unknown Command: `%s`\n", value)
    }

    return value, err
}
