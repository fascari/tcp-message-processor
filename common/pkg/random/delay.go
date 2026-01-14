package random

import (
	"math/rand/v2"
	"time"
)

// Delay returns a random duration between min and max (inclusive).
// If min equals max, returns min.
func Delay(min, max time.Duration) time.Duration {
	if min == max {
		return min
	}

	diff := max - min
	return min + time.Duration(rand.Int64N(int64(diff)))
}
