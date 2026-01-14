package noncegen

import (
	"crypto/rand"
	"encoding/hex"
)

// Generate returns a cryptographically secure random 32-character hex string.
// Uses 16 random bytes encoded as hexadecimal.
func Generate() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
