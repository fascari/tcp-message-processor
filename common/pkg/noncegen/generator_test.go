package noncegen_test

import (
	"testing"

	"tcp-message-processor/common/pkg/noncegen"

	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	t.Run("should generate 32 character hex string", func(t *testing.T) {
		result := noncegen.Generate()
		require.Len(t, result, 32)
	})

	t.Run("should generate unique nonces", func(t *testing.T) {
		nonce1 := noncegen.Generate()
		nonce2 := noncegen.Generate()
		require.NotEqual(t, nonce1, nonce2)
	})

	t.Run("should generate valid hexadecimal", func(t *testing.T) {
		result := noncegen.Generate()
		for _, char := range result {
			require.True(t, isHexChar(char), "character %c is not valid hex", char)
		}
	})

	t.Run("should generate multiple unique nonces", func(t *testing.T) {
		seen := make(map[string]bool)
		iterations := 100

		for i := 0; i < iterations; i++ {
			n := noncegen.Generate()
			require.False(t, seen[n], "duplicate nonce generated: %s", n)
			seen[n] = true
		}

		require.Len(t, seen, iterations)
	})
}

func isHexChar(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')
}
