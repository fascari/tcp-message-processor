package submit

import (
	"testing"

	"tcp-message-processor/common/pkg/tcp"

	"github.com/stretchr/testify/require"
)

func newSubmitPipe(t *testing.T) (client *tcp.Conn, server *tcp.Conn) {
	t.Helper()

	c1, c2, cleanup := tcp.NewPipe()
	t.Cleanup(cleanup)

	return c1, c2
}

func TestSubmit_Success(t *testing.T) {
	client, server := newSubmitPipe(t)
	submitter := New(0, 0)

	go func() {
		_, _ = server.Read() // consume request
		_ = server.Write(&tcp.Message{Result: true})
	}()

	err := submitter.Submit(client, 123, "nonce")

	require.NoError(t, err)
}

func TestSubmit_WriteError(t *testing.T) {
	client, server := newSubmitPipe(t)
	submitter := New(0, 0)

	_ = server.Close()

	err := submitter.Submit(client, 123, "nonce")

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to send submit request")
}

func TestSubmit_ReadError(t *testing.T) {
	client, server := newSubmitPipe(t)
	submitter := New(0, 0)

	go func() {
		_, _ = server.Read()
		_ = server.Close()
	}()

	err := submitter.Submit(client, 123, "nonce")

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read submit response")
}

func TestSubmit_ServerRejection(t *testing.T) {
	client, server := newSubmitPipe(t)
	submitter := New(0, 0)

	go func() {
		_, _ = server.Read()
		_ = server.Write(&tcp.Message{Error: "rate limit exceeded"})
	}()

	err := submitter.Submit(client, 123, "nonce")

	require.NoError(t, err)
}
