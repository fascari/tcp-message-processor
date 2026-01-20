package random

import (
	"math/rand/v2"
	"time"
)

// Delay returns a random duration between minDelay and maxDelay (inclusive).
// If minDelay equals maxDelay, returns minDelay.
func Delay(minDelay, maxDelay time.Duration) time.Duration {
	if minDelay == maxDelay {
		return minDelay
	}

	diff := maxDelay - minDelay
	return minDelay + time.Duration(rand.Int64N(int64(diff)))
}
