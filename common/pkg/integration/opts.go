package integration

import (
	"context"
)

type Option func(*Suite) error

func WithRabbitMQ() Option {
	return func(s *Suite) error {
		return s.startRabbitMQ()
	}
}

func WithServer() Option {
	return func(s *Suite) error {
		return s.startServer()
	}
}

func WithDatabase() Option {
	return func(s *Suite) error {
		if err := s.startPostgres(); err != nil {
			return err
		}

		ctx := context.Background()
		db, err := s.SetupPostgresWithMigrations(ctx)
		if err != nil {
			return err
		}
		s.db = db
		return nil
	}
}
