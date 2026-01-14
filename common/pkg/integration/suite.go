package integration

import (
	"context"
	"os/exec"

	"tcp-message-processor/common/pkg/env"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

type Suite struct {
	ctx               context.Context
	postgresContainer *postgres.PostgresContainer
	rabbitmqContainer *rabbitmq.RabbitMQContainer
	serverCmd         *exec.Cmd
	ServerHost        string
	ServerPort        string
}

func NewSuite() *Suite {
	return &Suite{
		ctx:        context.Background(),
		ServerHost: env.Lookup("SERVER_HOST", "localhost"),
		ServerPort: env.Lookup("SERVER_PORT", "8888"),
	}
}

func (s *Suite) Setup() error {
	if err := s.startContainers(); err != nil {
		return err
	}

	if err := s.startServer(); err != nil {
		s.Teardown()
		return err
	}

	return nil
}

func (s *Suite) Teardown() {
	s.stopServer()
	s.stopContainers()
}
