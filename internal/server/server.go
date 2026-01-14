package server

import (
	"context"
	"time"

	"tcp-message-processor/internal/session"
)

type (
	Server struct {
		sessions    *session.Store
		publisher   Publisher
		broadcaster *broadcaster
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

	JobParams struct {
		JobID       int64  `json:"job_id"`
		ServerNonce string `json:"server_nonce"`
	}
)

func (p JobParams) ToMap() map[string]any {
	return map[string]any{
		"job_id":       p.JobID,
		"server_nonce": p.ServerNonce,
	}
}

func New(sessions *session.Store, publisher Publisher, broadcastInterval int) Server {
	return Server{
		sessions:    sessions,
		publisher:   publisher,
		broadcaster: newBroadcaster(sessions, broadcastInterval),
	}
}

func (s *Server) Start(ctx context.Context) {
	s.broadcaster.start(ctx)
}

func (s *Server) Stop() {
	s.broadcaster.stop()
}
