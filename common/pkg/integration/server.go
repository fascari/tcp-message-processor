package integration

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"tcp-message-processor/common/pkg/closer"
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
	cwd, _ := os.Getwd()
	serverDir := findServerDir(cwd)

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

	rabbitmqURL := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		s.config.RabbitMQ.User,
		s.config.RabbitMQ.Password,
		rabbitmqHost,
		rabbitmqPort.Port())

	return append(os.Environ(),
		fmt.Sprintf("DB_HOST=%s", postgresHost),
		fmt.Sprintf("DB_PORT=%s", postgresPort.Port()),
		fmt.Sprintf("DB_USER=%s", s.config.DB.User),
		fmt.Sprintf("DB_PASSWORD=%s", s.config.DB.Password),
		fmt.Sprintf("DB_NAME=%s", s.config.DB.Name),
		fmt.Sprintf("DB_SSLMODE=%s", s.config.DB.SSLMode),
		fmt.Sprintf("RABBITMQ_URL=%s", rabbitmqURL),
		fmt.Sprintf("SERVER_HOST=%s", s.config.Server.Host),
		fmt.Sprintf("SERVER_PORT=%s", s.config.Server.Port),
		fmt.Sprintf("BROADCAST_INTERVAL_SECONDS=%s", s.config.Server.BroadcastIntervalSeconds),
	), nil
}

func (s *Suite) executeServer(serverDir string, envVars []string) error {
	s.serverCmd = exec.Command("go", "run", "cmd/server/main.go")
	s.serverCmd.Dir = serverDir
	s.serverCmd.Env = envVars
	s.serverCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err == nil {
		s.serverCmd.Stdout = devNull
		s.serverCmd.Stderr = devNull
	}

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
