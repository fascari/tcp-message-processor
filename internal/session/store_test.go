package session

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStore_Create(t *testing.T) {
	store := NewStore()

	session := store.Create("user1")

	require.NotNil(t, session)
	require.Equal(t, "user1", session.Username)
}

func TestStore_Find(t *testing.T) {
	store := NewStore()

	store.Create("user1")

	session, exists := store.Find("user1")
	require.True(t, exists)
	require.Equal(t, "user1", session.Username)

	_, exists = store.Find("nonexistent")
	require.False(t, exists)
}

func TestStore_Delete(t *testing.T) {
	store := NewStore()

	store.Create("user1")
	store.Delete("user1")

	_, exists := store.Find("user1")
	require.False(t, exists)
}

func TestStore_All(t *testing.T) {
	store := NewStore()

	store.Create("user1")
	store.Create("user2")
	store.Create("user3")

	sessions := store.All()
	require.Len(t, sessions, 3)
}

// TestStore_ConcurrentAccess validates that Store is thread-safe for concurrent writes.
// It launches 100 goroutines creating sessions simultaneously with different usernames (A-Z cycling).
// Uses rune('A' + i) to generate unique single-character usernames like "A", "B", "C", etc.
// Verifies all 100 sessions are created without race conditions or data loss.
func TestStore_ConcurrentAccess(t *testing.T) {
	store := NewStore()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Go(func() {
			username := string(rune('A' + i))
			store.Create(username)
		})
	}

	wg.Wait()

	sessions := store.All()
	require.Len(t, sessions, 100, "should create all sessions without race conditions when accessed concurrently")
}

// TestStore_ConcurrentReadWrite validates that Store is thread-safe for concurrent reads.
// It launches 200 goroutines (100 Find + 100 All) reading simultaneously and verifies
// that all operations return correct data without race conditions or corruption.
func TestStore_ConcurrentReadWrite(t *testing.T) {
	store := NewStore()
	store.Create("user1")

	var wg sync.WaitGroup
	successCount := 0
	mu := sync.Mutex{}

	for i := 0; i < 100; i++ {
		wg.Go(func() {
			session, exists := store.Find("user1")
			if exists && session.Username == "user1" {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		})

		wg.Go(func() {
			sessions := store.All()
			if len(sessions) == 1 {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		})
	}

	wg.Wait()

	require.Equal(t, 200, successCount, "should successfully read all concurrent operations without errors")
}
