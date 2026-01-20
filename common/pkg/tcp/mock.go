package tcp

import "net"

func NewPipe() (client *Conn, server *Conn, cleanup func()) {
	c1, c2 := net.Pipe()
	cleanup = func() {
		_ = c1.Close()
		_ = c2.Close()
	}
	return NewConn(c1), NewConn(c2), cleanup
}
