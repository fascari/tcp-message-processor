package handler

import (
	"context"
	"errors"
	"net"

	"tcp-message-processor/common/pkg/logger"
	"tcp-message-processor/common/pkg/tcp"

	"go.uber.org/zap"
)

func (s *Server) authorize(_ context.Context, conn net.Conn, msg tcp.Message) (string, error) {
	username, ok := msg.Params["username"].(string)
	if !ok || username == "" {
		s.sendError(conn, *msg.ID, "invalid username")
		return "", errors.New("invalid username in params")
	}

	s.sessions.Create(username)
	s.broadcaster.Register(username, conn)

	response := tcp.NewSuccessResponse(*msg.ID, true)
	if err := tcp.WriteMessage(conn, response); err != nil {
		return "", err
	}

	logger.Info("client authorized", zap.String("username", username))
	return username, nil
}
