package handler

import (
	"errors"

	"tcp-message-processor/common/pkg/logger"
	"tcp-message-processor/common/pkg/tcp"

	"go.uber.org/zap"
)

func (h *handler) authorize(conn *tcp.Conn, msg tcp.Message) (string, error) {
	username, ok := msg.Params["username"].(string)
	if !ok || username == "" {
		h.sendError(conn, *msg.ID, "invalid username")
		return "", errors.New("invalid username in params")
	}

	h.sessions.Create(username)
	h.broadcaster.Register(username, conn)

	response := tcp.SuccessResponse(*msg.ID, true)
	if err := conn.Write(&response); err != nil {
		return "", err
	}

	logger.Info("client authorized", zap.String("username", username))
	return username, nil
}
