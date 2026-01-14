package submit

import (
	"errors"
	"testing"

	"tcp-message-processor-client/internal/transport/mocks"
	"tcp-message-processor/common/pkg/tcp"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSubmit_Success(t *testing.T) {
	mockConn := mocks.NewMessenger(t)
	submitter := New(0, 0)

	mockConn.EXPECT().Write(mock.Anything).Return(nil).Once()
	mockConn.EXPECT().Read().Return(tcp.Message{Result: true}, nil).Once()

	err := submitter.Submit(mockConn, 123, "nonce")

	require.NoError(t, err)
}

func TestSubmit_WriteError(t *testing.T) {
	mockConn := mocks.NewMessenger(t)
	submitter := New(0, 0)

	mockConn.EXPECT().Write(mock.Anything).Return(errors.New("write failed")).Once()

	err := submitter.Submit(mockConn, 123, "nonce")

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to send submit request")
}

func TestSubmit_ReadError(t *testing.T) {
	mockConn := mocks.NewMessenger(t)
	submitter := New(0, 0)

	mockConn.EXPECT().Write(mock.Anything).Return(nil).Once()
	mockConn.EXPECT().Read().Return(tcp.Message{}, errors.New("read failed")).Once()

	err := submitter.Submit(mockConn, 123, "nonce")

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read submit response")
}

func TestSubmit_ServerRejection(t *testing.T) {
	mockConn := mocks.NewMessenger(t)
	submitter := New(0, 0)

	mockConn.EXPECT().Write(mock.Anything).Return(nil).Once()
	mockConn.EXPECT().Read().Return(tcp.Message{Error: "rate limit exceeded"}, nil).Once()

	err := submitter.Submit(mockConn, 123, "nonce")

	require.NoError(t, err)
}
