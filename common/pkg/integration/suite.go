package integration

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os/exec"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

type Suite struct {
	ctx               context.Context
	postgresContainer *postgres.PostgresContainer
	rabbitmqContainer *rabbitmq.RabbitMQContainer
	serverCmd         *exec.Cmd
	db                *sql.DB
	config            *Config
	ServerHost        string
	ServerPort        string
}

func NewSuite(opts ...Option) (*Suite, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load integration config: %w", err)
	}

	s := &Suite{
		ctx:        context.Background(),
		config:     cfg,
		ServerHost: cfg.Server.Host,
		ServerPort: cfg.Server.Port,
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			s.Teardown()
			return nil, err
		}
	}

	return s, nil
}

func (s *Suite) Teardown() {
	s.stopServer()
	s.stopContainers()
	if s.db != nil {
		_ = s.db.Close()
	}
}

func (s *Suite) DB() *sql.DB {
	return s.db
}

func (s *Suite) Connect() (net.Conn, error) {
	addr := net.JoinHostPort(s.ServerHost, s.ServerPort)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}
	return conn, nil
}
