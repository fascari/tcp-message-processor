package random_test

import (
	"testing"
	"time"

	"tcp-message-processor/common/pkg/random"

	"github.com/stretchr/testify/require"
)

func TestDelay(t *testing.T) {
	t.Run("should return min when min equals max", func(t *testing.T) {
		minDelay := 5 * time.Second
		maxDelay := 5 * time.Second

		result := random.Delay(minDelay, maxDelay)
		require.Equal(t, minDelay, result)
	})

	t.Run("should return value between min and max", func(t *testing.T) {
		minDelay := 1 * time.Second
		maxDelay := 10 * time.Second

		for i := 0; i < 100; i++ {
			result := random.Delay(minDelay, maxDelay)
			require.GreaterOrEqual(t, result, minDelay)
			require.LessOrEqual(t, result, maxDelay)
		}
	})

	t.Run("should generate different delays", func(t *testing.T) {
		minDelay := 1 * time.Second
		maxDelay := 60 * time.Second

		seen := make(map[time.Duration]bool)
		iterations := 50

		for i := 0; i < iterations; i++ {
			result := random.Delay(minDelay, maxDelay)
			seen[result] = true
		}

		require.Greater(t, len(seen), 1, "should generate at least 2 different values")
	})

	t.Run("should handle zero duration", func(t *testing.T) {
		minDelay := 0 * time.Second
		maxDelay := 0 * time.Second

		result := random.Delay(minDelay, maxDelay)
		require.Equal(t, time.Duration(0), result)
	})

	t.Run("should handle milliseconds", func(t *testing.T) {
		minDelay := 100 * time.Millisecond
		maxDelay := 200 * time.Millisecond

		result := random.Delay(minDelay, maxDelay)
		require.GreaterOrEqual(t, result, minDelay)
		require.LessOrEqual(t, result, maxDelay)
	})
}
