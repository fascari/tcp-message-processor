package random_test

import (
	"testing"
	"time"

	"tcp-message-processor/common/pkg/random"

	"github.com/stretchr/testify/require"
)

func TestDelay(t *testing.T) {
	t.Run("should return min when min equals max", func(t *testing.T) {
		min := 5 * time.Second
		max := 5 * time.Second

		result := random.Delay(min, max)
		require.Equal(t, min, result)
	})

	t.Run("should return value between min and max", func(t *testing.T) {
		min := 1 * time.Second
		max := 10 * time.Second

		for i := 0; i < 100; i++ {
			result := random.Delay(min, max)
			require.GreaterOrEqual(t, result, min)
			require.LessOrEqual(t, result, max)
		}
	})

	t.Run("should generate different delays", func(t *testing.T) {
		min := 1 * time.Second
		max := 60 * time.Second

		seen := make(map[time.Duration]bool)
		iterations := 50

		for i := 0; i < iterations; i++ {
			result := random.Delay(min, max)
			seen[result] = true
		}

		require.Greater(t, len(seen), 1, "should generate at least 2 different values")
	})

	t.Run("should handle zero duration", func(t *testing.T) {
		min := 0 * time.Second
		max := 0 * time.Second

		result := random.Delay(min, max)
		require.Equal(t, time.Duration(0), result)
	})

	t.Run("should handle milliseconds", func(t *testing.T) {
		min := 100 * time.Millisecond
		max := 200 * time.Millisecond

		result := random.Delay(min, max)
		require.GreaterOrEqual(t, result, min)
		require.LessOrEqual(t, result, max)
	})
}
