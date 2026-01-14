package jobs_test

import (
	"testing"

	"tcp-message-processor-client/internal/jobs"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestManager_Update(t *testing.T) {
	tests := []struct {
		name        string
		jobID       int64
		serverNonce string
	}{
		{
			name:        "should update with valid job data",
			jobID:       1,
			serverNonce: "abc123",
		},
		{
			name:        "should update with different job data",
			jobID:       999,
			serverNonce: "xyz789",
		},
		{
			name:        "should handle zero job_id",
			jobID:       0,
			serverNonce: "nonce",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := jobs.NewManager()
			manager.Update(tt.jobID, tt.serverNonce)

			gotJobID, gotNonce := manager.Current()
			require.Equal(t, tt.jobID, gotJobID)
			require.Equal(t, tt.serverNonce, gotNonce)
		})
	}
}

func TestManager_Current(t *testing.T) {
	t.Run("should return zero values when not initialized", func(t *testing.T) {
		manager := jobs.NewManager()

		jobID, nonce := manager.Current()
		require.Equal(t, int64(0), jobID)
		require.Equal(t, "", nonce)
	})

	t.Run("should return current job after update", func(t *testing.T) {
		manager := jobs.NewManager()
		manager.Update(42, "test-nonce")

		jobID, nonce := manager.Current()
		require.Equal(t, int64(42), jobID)
		require.Equal(t, "test-nonce", nonce)
	})
}

func TestManager_ConcurrentAccess(t *testing.T) {
	t.Run("should handle concurrent updates and reads safely", func(t *testing.T) {
		manager := jobs.NewManager()
		var wg errgroup.Group

		updates := 100

		for i := 0; i < updates; i++ {
			id := int64(i)
			wg.Go(func() error {
				manager.Update(id, "nonce")
				return nil
			})

			wg.Go(func() error {
				_, _ = manager.Current()
				return nil
			})
		}

		err := wg.Wait()
		require.NoError(t, err)

		jobID, nonce := manager.Current()
		require.NotEqual(t, int64(0), jobID)
		require.Equal(t, "nonce", nonce)
	})
}
