package session

import "sync"

type Store struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

func NewStore() *Store {
	return &Store{
		sessions: make(map[string]*Session),
	}
}

func (s *Store) Create(username string) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := New(username)
	s.sessions[username] = session
	return session
}

func (s *Store) Find(username string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[username]
	return session, exists
}

func (s *Store) Delete(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, username)
}

func (s *Store) All() []*Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}
