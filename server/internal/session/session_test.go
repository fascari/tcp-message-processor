package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSession_UpdateJob(t *testing.T) {
	session := New("testuser")

	session.UpdateJob(1, "nonce1")

	require.Equal(t, int64(1), session.CurrentJobID)
	require.Equal(t, "nonce1", session.CurrentNonce())
	require.Equal(t, "nonce1", session.JobHistory[1])
	require.False(t, session.NonceUpdatedAt.IsZero())
}

func TestSession_RecordSubmission(t *testing.T) {
	session := New("testuser")

	session.RecordSubmission("client_nonce_1")

	require.True(t, session.IsDuplicateNonce("client_nonce_1"))
	require.False(t, session.IsDuplicateNonce("client_nonce_2"))
}

func TestSession_AllowSubmission(t *testing.T) {
	session := New("testuser")

	require.True(t, session.AllowSubmission(), "should allow first submission")
	require.False(t, session.AllowSubmission(), "should deny second submission when within rate limit window")

	time.Sleep(1100 * time.Millisecond)

	require.True(t, session.AllowSubmission(), "should allow submission when 1 second has passed")
}

func TestSession_IsDuplicateNonce(t *testing.T) {
	session := New("testuser")

	require.False(t, session.IsDuplicateNonce("nonce1"))

	session.RecordSubmission("nonce1")

	require.True(t, session.IsDuplicateNonce("nonce1"))
}

func TestSession_ValidateJob(t *testing.T) {
	session := New("testuser")
	session.UpdateJob(1, "server_nonce_1")

	validation := session.ValidateJob(1)
	require.True(t, validation.Exists)
	require.False(t, validation.IsExpired)
	require.True(t, validation.NonceMatches)
	require.Equal(t, "server_nonce_1", validation.CurrentNonce)

	validation = session.ValidateJob(999)
	require.False(t, validation.Exists)
}

func TestSession_JobExpiration(t *testing.T) {
	session := New("testuser")

	session.UpdateJob(1, "nonce1")
	validation := session.ValidateJob(1)
	require.False(t, validation.IsExpired, "should not be expired when job is current")

	session.UpdateJob(2, "nonce2")
	validation = session.ValidateJob(1)
	require.True(t, validation.IsExpired, "should be expired when newer job exists")

	validation = session.ValidateJob(2)
	require.False(t, validation.IsExpired, "should not be expired when job is current")

	session.UpdateJob(5, "nonce5")
	validation = session.ValidateJob(1)
	require.True(t, validation.IsExpired, "should be expired when much newer job exists")

	validation = session.ValidateJob(2)
	require.True(t, validation.IsExpired, "should be expired when newer job exists")

	validation = session.ValidateJob(5)
	require.False(t, validation.IsExpired, "should not be expired when job is current")

	validation = session.ValidateJob(999)
	require.False(t, validation.IsExpired, "should not be expired when job never existed")

	validation = session.ValidateJob(4)
	require.False(t, validation.IsExpired, "should not be expired when job never existed")
}
