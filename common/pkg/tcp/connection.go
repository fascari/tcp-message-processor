package tcp

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
)

const (
	// maxMessageSize limits messages to 1 MiB (1 << 20 = 2^20 bytes).
	// Prevents DoS attacks where clients send huge size values causing OOM.
	maxMessageSize = 1 << 20

	// bufferSize sets I/O buffers to 8 KiB, matching typical memory page size.
	// Reduces syscalls while keeping memory usage reasonable.
	bufferSize = 8192
)

type Conn struct {
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
	writeMux sync.Mutex
}

func NewConn(conn net.Conn) *Conn {
	return &Conn{
		conn:   conn,
		reader: bufio.NewReaderSize(conn, bufferSize),
		writer: bufio.NewWriterSize(conn, bufferSize),
	}
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) Conn() net.Conn {
	return c.conn
}

func (c *Conn) Read() (Message, error) {
	var length uint32
	if err := binary.Read(c.reader, binary.BigEndian, &length); err != nil {
		return Message{}, err
	}

	if length > maxMessageSize {
		return Message{}, fmt.Errorf("message too large: %d bytes (max: %d)", length, maxMessageSize)
	}

	payload := make([]byte, length)
	if _, err := io.ReadFull(c.reader, payload); err != nil {
		return Message{}, err
	}

	var msg Message
	if err := json.Unmarshal(payload, &msg); err != nil {
		return Message{}, err
	}

	return msg, nil
}

func (c *Conn) Write(msg *Message) error {
	c.writeMux.Lock()
	defer c.writeMux.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	length := uint32(len(data))
	if err := binary.Write(c.writer, binary.BigEndian, length); err != nil {
		return err
	}

	if _, err := c.writer.Write(data); err != nil {
		return err
	}

	return c.writer.Flush()
}
