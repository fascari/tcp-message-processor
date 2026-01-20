package transport

import (
	"fmt"
	"net"

	"tcp-message-processor/common/pkg/tcp"
)

func Connect(host, port string) (*tcp.Conn, error) {
	addr := net.JoinHostPort(host, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	return tcp.NewConn(conn), nil
}
