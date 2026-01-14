package ratelimit

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestLimiter_Allow(t *testing.T) {
	limiter := NewSecondLimiter()

	require.True(t, limiter.Allow("user1"), "should allow first submission")
	require.False(t, limiter.Allow("user1"), "should deny second submission when within 1 second")

	time.Sleep(1100 * time.Millisecond)

	require.True(t, limiter.Allow("user1"), "should allow submission when 1 second has passed")
}

func TestLimiter_AllowMultipleUsers(t *testing.T) {
	limiter := NewSecondLimiter()

	require.True(t, limiter.Allow("user1"))
	require.True(t, limiter.Allow("user2"))

	require.False(t, limiter.Allow("user1"))
	require.False(t, limiter.Allow("user2"))
}

func TestLimiter_ConcurrentAccess(t *testing.T) {
	limiter := NewSecondLimiter()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Go(func() {
			key := string(rune('A' + (i % 26)))
			limiter.Allow(key)
		})
	}

	wg.Wait()

	require.NotEmpty(t, limiter.limiters, "should record limiters when accessed concurrently")
}

func TestLimiter_CustomRate(t *testing.T) {
	limiter := New(rate.Limit(2), 2)

	require.True(t, limiter.Allow("user1"))
	require.True(t, limiter.Allow("user1"))
	require.False(t, limiter.Allow("user1"), "should deny third request when limit is 2 req/s")

	time.Sleep(1100 * time.Millisecond)

	require.True(t, limiter.Allow("user1"), "should allow request when window refills")
}

func TestLimiter_BurstAllowance(t *testing.T) {
	limiter := New(rate.Limit(1), 3)

	require.True(t, limiter.Allow("user1"))
	require.True(t, limiter.Allow("user1"))
	require.True(t, limiter.Allow("user1"))
	require.False(t, limiter.Allow("user1"), "should deny fourth request when burst is 3")
}
