package hash_test

import (
	"testing"

	"tcp-message-processor/common/pkg/hash"

	"github.com/stretchr/testify/require"
)

func TestSHA256(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "should return correct hash when input is empty string",
			input:    "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "should return correct hash when input is numeric",
			input:    "123456",
			expected: "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92",
		},
		{
			name:     "should return correct hash when input contains letters and numbers",
			input:    "abc123",
			expected: "6ca13d52ca70c883e0f0bb101e425a89e8624de51db2d2392593af6a84118090",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hash.SHA256(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}
