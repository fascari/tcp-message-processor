package handler

import (
	"context"
	"time"

	"tcp-message-processor/internal/broadcaster"
	"tcp-message-processor/internal/session"
)

//go:generate mockery --all

type (
	Server struct {
		sessions    *session.Store
		publisher   Publisher
		broadcaster broadcaster.Broadcaster
	}

	Publisher interface {
		Publish(ctx context.Context, event Event) error
	}

	Event struct {
		Username    string
		JobID       int64
		ClientNonce string
		Timestamp   time.Time
	}
)

func New(sessions *session.Store, publisher Publisher, broadcastInterval int) *Server {
	return &Server{
		sessions:    sessions,
		publisher:   publisher,
		broadcaster: broadcaster.New(sessions, broadcastInterval),
	}
}

func (s *Server) Start() {
	s.broadcaster.Start()
}

func (s *Server) Stop() {
	s.broadcaster.Stop()
}
