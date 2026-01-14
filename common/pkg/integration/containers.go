package integration

import (
	"fmt"
	"time"

	"tcp-message-processor/common/pkg/env"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"
)

func (s *Suite) startContainers() error {
	if err := s.startPostgres(); err != nil {
		return err
	}

	if err := s.startRabbitMQ(); err != nil {
		return err
	}

	return nil
}

func (s *Suite) startPostgres() error {
	dbName := env.Lookup("DB_NAME", "tcpprocessor")
	dbUser := env.Lookup("DB_USER", "tcpuser")
	dbPassword := env.Lookup("DB_PASSWORD", "tcppass")

	container, err := postgres.Run(s.ctx,
		"postgres:17",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		return fmt.Errorf("failed to start postgres: %w", err)
	}

	s.postgresContainer = container
	return nil
}

func (s *Suite) startRabbitMQ() error {
	container, err := rabbitmq.Run(s.ctx,
		"rabbitmq:4.0-management",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Server startup complete").
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		return fmt.Errorf("failed to start rabbitmq: %w", err)
	}

	s.rabbitmqContainer = container
	return nil
}
