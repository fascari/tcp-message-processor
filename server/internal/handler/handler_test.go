package handler_test

import (
	"context"
	"testing"
	"time"

	"tcp-message-processor/common/pkg/closer"
	"tcp-message-processor/common/pkg/hash"
	"tcp-message-processor/common/pkg/noncegen"
	"tcp-message-processor/common/pkg/tcp"
	handlerpkg "tcp-message-processor/internal/handler"
	"tcp-message-processor/internal/handler/mocks"
	"tcp-message-processor/internal/session"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupHandlerWithConnection(t *testing.T) (handler handlerpkg.Handler, serverConn *tcp.Conn, clientConn *tcp.Conn) {
	sessions := session.NewStore()
	publisher := mocks.NewPublisher(t)
	publisher.EXPECT().Publish(
		mock.Anything,
		mock.Anything,
	).Return(nil).Maybe()

	h := handlerpkg.New(sessions, publisher, 1)

	serverConn, clientConn, cleanup := tcp.NewPipe()
	t.Cleanup(cleanup)

	ctx := context.Background()
	go h.Handle(ctx, serverConn.Conn())

	return h, serverConn, clientConn
}

func authenticate(t *testing.T, tcpConn *tcp.Conn, username string) {
	authMsg := tcp.AuthorizeParams{Username: username}.ToMessage(1)
	require.NoError(t, tcpConn.Write(&authMsg))

	response, err := tcpConn.Read()
	require.NoError(t, err)
	require.Empty(t, response.Error)
	require.True(t, response.Result.(bool))
}

func waitForJob(t *testing.T, tcpConn *tcp.Conn) (int64, string) {
	jobMsg, err := tcpConn.Read()
	require.NoError(t, err)
	require.True(t, jobMsg.IsJob())

	jobID := int64(jobMsg.Params["job_id"].(float64))
	serverNonce := jobMsg.Params["server_nonce"].(string)

	return jobID, serverNonce
}

func readResponse(t *testing.T, conn *tcp.Conn) tcp.Message {
	for {
		msg, err := conn.Read()
		require.NoError(t, err)

		if msg.IsJob() {
			continue
		}

		return msg
	}
}

func submitAndReadResponse(t *testing.T, tcpConn *tcp.Conn, jobID int64, clientNonce, result string, msgID int64) tcp.Message {
	submitMsg := tcp.SubmitParams{
		JobID:       jobID,
		ClientNonce: clientNonce,
		Result:      result,
	}.ToMessage(msgID)

	require.NoError(t, tcpConn.Write(&submitMsg))
	return readResponse(t, tcpConn)
}

func setupTest(t *testing.T) *tcp.Conn {
	h, _, clientConn := setupHandlerWithConnection(t)
	t.Cleanup(func() {
		closer.Close(clientConn, "close client connection")
		h.Stop()
	})

	h.Start()
	authenticate(t, clientConn, "testuser")

	return clientConn
}

func TestHandler_ErrorConditions(t *testing.T) {
	t.Run("should return 'Duplicate submission' when nonce is reused", func(t *testing.T) {
		clientConn := setupTest(t)

		jobID, serverNonce := waitForJob(t, clientConn)
		clientNonce := noncegen.Generate()
		result := hash.SHA256(serverNonce + clientNonce)

		response1 := submitAndReadResponse(t, clientConn, jobID, clientNonce, result, 2)
		require.Empty(t, response1.Error)

		time.Sleep(1100 * time.Millisecond)

		response2 := submitAndReadResponse(t, clientConn, jobID, clientNonce, result, 3)
		require.Equal(t, "duplicate submission", response2.Error)
		require.False(t, response2.Result.(bool))
	})

	t.Run("should return 'Task expired' when job_id is old", func(t *testing.T) {
		clientConn := setupTest(t)

		job1ID, job1Nonce := waitForJob(t, clientConn)
		job2ID, _ := waitForJob(t, clientConn)
		require.Greater(t, job2ID, job1ID)

		clientNonce := noncegen.Generate()
		result := hash.SHA256(job1Nonce + clientNonce)

		response := submitAndReadResponse(t, clientConn, job1ID, clientNonce, result, 2)
		require.Equal(t, "task expired", response.Error)
		require.False(t, response.Result.(bool))
	})

	t.Run("should return 'Task does not exist' when job_id is invalid", func(t *testing.T) {
		clientConn := setupTest(t)

		_, serverNonce := waitForJob(t, clientConn)
		clientNonce := noncegen.Generate()
		result := hash.SHA256(serverNonce + clientNonce)

		response := submitAndReadResponse(t, clientConn, 999, clientNonce, result, 2)
		require.Equal(t, "task does not exist", response.Error)
		require.False(t, response.Result.(bool))
	})

	t.Run("should return 'Invalid result' when SHA256 is wrong", func(t *testing.T) {
		clientConn := setupTest(t)

		jobID, _ := waitForJob(t, clientConn)
		clientNonce := noncegen.Generate()

		response := submitAndReadResponse(t, clientConn, jobID, clientNonce, "wrong_sha256_hash", 2)
		require.Equal(t, "invalid result", response.Error)
		require.False(t, response.Result.(bool))
	})
}
