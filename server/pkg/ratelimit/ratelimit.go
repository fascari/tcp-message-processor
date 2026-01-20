package ratelimit

import (
	"sync"

	"golang.org/x/time/rate"
)

type Limiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	rps      rate.Limit
	burst    int
}

func New(rps rate.Limit, burst int) *Limiter {
	return &Limiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rps,
		burst:    burst,
	}
}

func NewSecondLimiter() *Limiter {
	return New(rate.Limit(1), 1)
}

func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(l.rps, l.burst)
		l.limiters[key] = limiter
	}

	return limiter.Allow()
}
