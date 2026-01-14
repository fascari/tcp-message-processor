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
	require.Equal(t, "nonce1", session.CurrentNonce)
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

func TestSession_ValidateJobNonce(t *testing.T) {
	session := New("testuser")
	session.UpdateJob(1, "server_nonce_1")

	require.True(t, session.ValidateJobNonce(1, "server_nonce_1"))
	require.False(t, session.ValidateJobNonce(1, "wrong_nonce"))
	require.False(t, session.ValidateJobNonce(999, "server_nonce_1"))
}

func TestSession_IsJobExpired(t *testing.T) {
	session := New("testuser")

	session.UpdateJob(1, "nonce1")
	require.False(t, session.IsJobExpired(1), "should not be expired when job is current")

	session.UpdateJob(2, "nonce2")
	require.True(t, session.IsJobExpired(1), "should be expired when newer job exists")
	require.False(t, session.IsJobExpired(2), "should not be expired when job is current")

	session.UpdateJob(5, "nonce5")
	require.True(t, session.IsJobExpired(1), "should be expired when much newer job exists")
	require.True(t, session.IsJobExpired(2), "should be expired when newer job exists")
	require.False(t, session.IsJobExpired(5), "should not be expired when job is current")
	require.False(t, session.IsJobExpired(999), "should not be expired when job is from future")
	require.False(t, session.IsJobExpired(4), "should not be expired when job never existed")
}
