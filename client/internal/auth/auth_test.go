package auth

import (
	"testing"

	"tcp-message-processor/common/pkg/tcp"

	"github.com/stretchr/testify/require"
)

func newAuthPipe(t *testing.T) (client *tcp.Conn, server *tcp.Conn) {
	t.Helper()

	c1, c2, cleanup := tcp.NewPipe()
	t.Cleanup(cleanup)

	return c1, c2
}

func TestAuthorize_Success(t *testing.T) {
	client, server := newAuthPipe(t)
	authenticator := New("testuser")

	go func() {
		_, _ = server.Read() // consume request
		_ = server.Write(&tcp.Message{Result: true})
	}()

	err := authenticator.Authorize(client)

	require.NoError(t, err)
}

func TestAuthorize_WriteError(t *testing.T) {
	client, server := newAuthPipe(t)
	authenticator := New("testuser")

	_ = server.Close() // force write failure on client side

	err := authenticator.Authorize(client)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to send authorize request")
}

func TestAuthorize_ReadError(t *testing.T) {
	client, server := newAuthPipe(t)
	authenticator := New("testuser")

	go func() {
		_, _ = server.Read() // consume request
		_ = server.Close()   // no response
	}()

	err := authenticator.Authorize(client)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read authorize response")
}

func TestAuthorize_ServerError(t *testing.T) {
	client, server := newAuthPipe(t)
	authenticator := New("testuser")

	go func() {
		_, _ = server.Read()
		_ = server.Write(&tcp.Message{Error: "invalid credentials"})
	}()

	err := authenticator.Authorize(client)

	require.Error(t, err)
	require.Contains(t, err.Error(), "authorization failed")
}

func TestAuthorize_ResultFalse(t *testing.T) {
	client, server := newAuthPipe(t)
	authenticator := New("testuser")

	go func() {
		_, _ = server.Read()
		_ = server.Write(&tcp.Message{Result: false})
	}()

	err := authenticator.Authorize(client)

	require.Error(t, err)
	require.Contains(t, err.Error(), "authorization rejected")
}

func TestAuthorize_InvalidResult(t *testing.T) {
	client, server := newAuthPipe(t)
	authenticator := New("testuser")

	go func() {
		_, _ = server.Read()
		_ = server.Write(&tcp.Message{Result: "not a bool"})
	}()

	err := authenticator.Authorize(client)

	require.Error(t, err)
	require.Contains(t, err.Error(), "authorization rejected")
}
