package closer

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockCloser struct {
	closeErr error
	closed   bool
}

func (m *mockCloser) Close() error {
	m.closed = true
	return m.closeErr
}

func TestClose(t *testing.T) {
	tests := []struct {
		name         string
		closer       io.Closer
		msg          string
		expectClosed bool
		setupCloser  func() io.Closer
	}{
		{
			name:         "should close successfully when no error occurs",
			msg:          "test message",
			expectClosed: true,
			setupCloser: func() io.Closer {
				return &mockCloser{}
			},
		},
		{
			name:         "should close and log error when close returns error",
			msg:          "test message",
			expectClosed: true,
			setupCloser: func() io.Closer {
				return &mockCloser{closeErr: errors.New("close failed")}
			},
		},
		{
			name:         "should not panic when closer is nil",
			msg:          "test message",
			expectClosed: false,
			setupCloser: func() io.Closer {
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			closer := tt.setupCloser()

			Close(closer, tt.msg)

			if tt.expectClosed {
				mock, ok := closer.(*mockCloser)
				require.True(t, ok, "expected mockCloser type")
				require.True(t, mock.closed, "expected Close to be called")
			}
		})
	}
}
