package integration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"
)

func (s *Suite) startPostgres() error {
	container, err := postgres.Run(s.ctx,
		"postgres:17",
		postgres.WithDatabase(s.config.DB.Name),
		postgres.WithUsername(s.config.DB.User),
		postgres.WithPassword(s.config.DB.Password),
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

func (s *Suite) PostgresContainer() *postgres.PostgresContainer {
	return s.postgresContainer
}

func (s *Suite) PostgresConnectionString(ctx context.Context, opts ...string) (string, error) {
	if s.postgresContainer == nil {
		return "", errors.New("postgres container not started")
	}
	return s.postgresContainer.ConnectionString(ctx, opts...)
}

func (s *Suite) applySchema(ctx context.Context, db interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}) error {
	if s.postgresContainer == nil {
		return errors.New("postgres container not started")
	}

	migrationPath, err := findMigrationFile()
	if err != nil {
		return fmt.Errorf("failed to find migration file: %w", err)
	}

	migration, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	_, err = db.ExecContext(ctx, string(migration))
	if err != nil {
		return fmt.Errorf("failed to apply migration: %w", err)
	}

	return nil
}

func findMigrationFile() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	current := cwd
	for i := 0; i < 10; i++ {
		migrationPath := filepath.Join(current, "server", "db", "migrations", "001_initial_schema.sql")
		if _, err := os.Stat(migrationPath); err == nil {
			return migrationPath, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", errors.New("migration file not found")
}

func (s *Suite) SetupPostgresWithMigrations(ctx context.Context) (*sql.DB, error) {
	connStr, err := s.PostgresConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := s.applySchema(ctx, db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return db, nil
}
