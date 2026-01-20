package auth

import (
	"fmt"

	"tcp-message-processor/common/pkg/logger"
	"tcp-message-processor/common/pkg/tcp"

	"go.uber.org/zap"
)

//go:generate mockery --all

type (
	Authenticator struct {
		username string
	}
)

func New(username string) Authenticator {
	return Authenticator{username: username}
}

func (a *Authenticator) Authorize(conn *tcp.Conn) error {
	id := int64(1)
	params := tcp.AuthorizeParams{
		Username: a.username,
	}
	msg := params.ToMessage(id)

	if err := conn.Write(&msg); err != nil {
		return fmt.Errorf("failed to send authorize request: %w", err)
	}

	response, err := conn.Read()
	if err != nil {
		return fmt.Errorf("failed to read authorize response: %w", err)
	}

	if response.Error != "" {
		return fmt.Errorf("authorization failed: %s", response.Error)
	}

	if result, ok := response.Result.(bool); !ok || !result {
		return fmt.Errorf("authorization rejected")
	}

	logger.Info("authenticated successfully", zap.String("username", a.username))
	return nil
}
