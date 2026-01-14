package integration

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"tcp-message-processor/common/pkg/closer"
	"tcp-message-processor/common/pkg/env"
)

func (s *Suite) startServer() error {
	serverDir, err := resolveServerDir()
	if err != nil {
		return err
	}

	envVars, err := s.buildServerEnv()
	if err != nil {
		return err
	}

	if err := s.executeServer(serverDir, envVars); err != nil {
		return err
	}

	return s.waitForServer(30 * time.Second)
}

func resolveServerDir() (string, error) {
	serverDir := env.Lookup("SERVER_DIR", "")
	if serverDir == "" {
		cwd, _ := os.Getwd()
		serverDir = findServerDir(cwd)
	}

	absServerDir, err := filepath.Abs(serverDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve server directory: %w", err)
	}

	mainPath := filepath.Join(absServerDir, "cmd", "server", "main.go")
	if _, err := os.Stat(mainPath); err != nil {
		return "", fmt.Errorf("server main.go not found at %s: %w", mainPath, err)
	}

	return absServerDir, nil
}

func (s *Suite) buildServerEnv() ([]string, error) {
	postgresHost, err := s.postgresContainer.Host(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get postgres host: %w", err)
	}

	postgresPort, err := s.postgresContainer.MappedPort(s.ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get postgres port: %w", err)
	}

	rabbitmqHost, err := s.rabbitmqContainer.Host(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get rabbitmq host: %w", err)
	}

	rabbitmqPort, err := s.rabbitmqContainer.MappedPort(s.ctx, "5672")
	if err != nil {
		return nil, fmt.Errorf("failed to get rabbitmq port: %w", err)
	}

	dbHost := env.Lookup("DB_HOST", postgresHost)
	dbPort := env.Lookup("DB_PORT", postgresPort.Port())
	dbUser := env.Lookup("DB_USER", "tcpuser")
	dbPassword := env.Lookup("DB_PASSWORD", "tcppass")
	dbName := env.Lookup("DB_NAME", "tcpprocessor")
	dbSSLMode := env.Lookup("DB_SSLMODE", "disable")

	rabbitmqUser := env.Lookup("RABBITMQ_USER", "guest")
	rabbitmqPassword := env.Lookup("RABBITMQ_PASSWORD", "guest")
	rabbitmqURL := fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitmqUser, rabbitmqPassword, rabbitmqHost, rabbitmqPort.Port())

	serverHost := env.Lookup("SERVER_HOST", s.ServerHost)
	serverPort := env.Lookup("SERVER_PORT", s.ServerPort)
	broadcastInterval := env.Lookup("BROADCAST_INTERVAL_SECONDS", "30")

	return append(os.Environ(),
		fmt.Sprintf("DB_HOST=%s", dbHost),
		fmt.Sprintf("DB_PORT=%s", dbPort),
		fmt.Sprintf("DB_USER=%s", dbUser),
		fmt.Sprintf("DB_PASSWORD=%s", dbPassword),
		fmt.Sprintf("DB_NAME=%s", dbName),
		fmt.Sprintf("DB_SSLMODE=%s", dbSSLMode),
		fmt.Sprintf("RABBITMQ_URL=%s", rabbitmqURL),
		fmt.Sprintf("SERVER_HOST=%s", serverHost),
		fmt.Sprintf("SERVER_PORT=%s", serverPort),
		fmt.Sprintf("BROADCAST_INTERVAL_SECONDS=%s", broadcastInterval),
	), nil
}

func (s *Suite) executeServer(serverDir string, envVars []string) error {
	s.serverCmd = exec.Command("go", "run", "cmd/server/main.go")
	s.serverCmd.Dir = serverDir
	s.serverCmd.Env = envVars

	if err := s.serverCmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func (s *Suite) waitForServer(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if s.isServerReady() {
			time.Sleep(2 * time.Second)
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("server did not start within %v", timeout)
}

func (s *Suite) isServerReady() bool {
	conn, err := net.Dial("tcp", net.JoinHostPort(s.ServerHost, s.ServerPort))
	if err != nil {
		return false
	}
	closer.Close(conn, "failed to close connection")
	return true
}

func findServerDir(startDir string) string {
	current := startDir
	for i := 0; i < 10; i++ {
		serverPath := filepath.Join(current, "server")
		if _, err := os.Stat(filepath.Join(serverPath, "cmd", "server", "main.go")); err == nil {
			return serverPath
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return filepath.Join(startDir, "server")
}
