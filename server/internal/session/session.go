package session

import (
	"sync"
	"time"

	"tcp-message-processor/pkg/ratelimit"
)

type Session struct {
	mu             sync.RWMutex
	Username       string
	CurrentJobID   int64
	currentNonce   string
	NonceUpdatedAt time.Time
	Submissions    map[string]time.Time
	JobHistory     map[int64]string
	limiter        *ratelimit.Limiter
}

func New(username string) *Session {
	return &Session{
		Username:    username,
		Submissions: make(map[string]time.Time),
		JobHistory:  make(map[int64]string),
		limiter:     ratelimit.NewSecondLimiter(),
	}
}

func (s *Session) UpdateJob(jobID int64, nonce string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.CurrentJobID = jobID
	s.currentNonce = nonce
	s.NonceUpdatedAt = time.Now()
	s.JobHistory[jobID] = nonce
}

func (s *Session) RecordSubmission(clientNonce string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Submissions[clientNonce] = time.Now()
}

func (s *Session) IsDuplicateNonce(clientNonce string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.Submissions[clientNonce]
	return exists
}

func (s *Session) ValidateJobNonce(jobID int64, expectedNonce string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nonce, exists := s.JobHistory[jobID]
	if !exists {
		return false
	}
	return nonce == expectedNonce
}

func (s *Session) IsJobExpired(jobID int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.JobHistory[jobID]
	if !exists {
		return false
	}
	return jobID < s.CurrentJobID
}

func (s *Session) CurrentNonce() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentNonce
}

func (s *Session) AllowSubmission() bool {
	return s.limiter.Allow(s.Username)
}
