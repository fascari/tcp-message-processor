package transport

import (
	"fmt"
	"net"

	"tcp-message-processor/common/pkg/tcp"
)

//go:generate mockery --all

type (
	Connection struct {
		conn net.Conn
	}

	Messenger interface {
		Read() (tcp.Message, error)
		Write(tcp.Message) error
	}
)

func Connect(host, port string) (Connection, error) {
	addr := net.JoinHostPort(host, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return Connection{}, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	return Connection{conn: conn}, nil
}

func (c Connection) Read() (tcp.Message, error) {
	return tcp.ReadMessage(c.conn)
}

func (c Connection) Write(msg tcp.Message) error {
	return tcp.WriteMessage(c.conn, &msg)
}

func (c Connection) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
