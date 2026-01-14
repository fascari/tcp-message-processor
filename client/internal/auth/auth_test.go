package auth

import (
	"errors"
	"testing"

	"tcp-message-processor-client/internal/transport/mocks"
	"tcp-message-processor/common/pkg/tcp"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthorize_Success(t *testing.T) {
	mockConn := mocks.NewMessenger(t)
	authenticator := New("testuser")

	mockConn.EXPECT().Write(mock.Anything).Return(nil).Once()
	mockConn.EXPECT().Read().Return(tcp.Message{Result: true}, nil).Once()

	err := authenticator.Authorize(mockConn)

	require.NoError(t, err)
}

func TestAuthorize_WriteError(t *testing.T) {
	mockConn := mocks.NewMessenger(t)
	authenticator := New("testuser")

	mockConn.EXPECT().Write(mock.Anything).Return(errors.New("connection reset")).Once()

	err := authenticator.Authorize(mockConn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to send authorize request")
}

func TestAuthorize_ReadError(t *testing.T) {
	mockConn := mocks.NewMessenger(t)
	authenticator := New("testuser")

	mockConn.EXPECT().Write(mock.Anything).Return(nil).Once()
	mockConn.EXPECT().Read().Return(tcp.Message{}, errors.New("read timeout")).Once()

	err := authenticator.Authorize(mockConn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read authorize response")
}

func TestAuthorize_ServerError(t *testing.T) {
	mockConn := mocks.NewMessenger(t)
	authenticator := New("testuser")

	mockConn.EXPECT().Write(mock.Anything).Return(nil).Once()
	mockConn.EXPECT().Read().Return(tcp.Message{Error: "invalid credentials"}, nil).Once()

	err := authenticator.Authorize(mockConn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "authorization failed")
}

func TestAuthorize_ResultFalse(t *testing.T) {
	mockConn := mocks.NewMessenger(t)
	authenticator := New("testuser")

	mockConn.EXPECT().Write(mock.Anything).Return(nil).Once()
	mockConn.EXPECT().Read().Return(tcp.Message{Result: false}, nil).Once()

	err := authenticator.Authorize(mockConn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "authorization rejected")
}

func TestAuthorize_InvalidResult(t *testing.T) {
	mockConn := mocks.NewMessenger(t)
	authenticator := New("testuser")

	mockConn.EXPECT().Write(mock.Anything).Return(nil).Once()
	mockConn.EXPECT().Read().Return(tcp.Message{Result: "not a bool"}, nil).Once()

	err := authenticator.Authorize(mockConn)

	require.Error(t, err)
	require.Contains(t, err.Error(), "authorization rejected")
}
