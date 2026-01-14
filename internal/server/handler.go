package server

import (
	"context"
	"net"

	"tcp-message-processor/pkg/closer"
	"tcp-message-processor/pkg/logger"
	"tcp-message-processor/pkg/tcp"

	"go.uber.org/zap"
)

func (s *Server) Handle(ctx context.Context, conn net.Conn) {
	defer closer.Close(conn, "failed to close connection")

	var username string
	defer func() {
		if username != "" {
			s.broadcaster.unregister(username)
			logger.Info("client disconnected", zap.String("username", username))
		}
	}()

	for {
		msg, err := tcp.ReadMessage(conn)
		if err != nil {
			if username != "" {
				logger.Warn("failed to read message", zap.Error(err), zap.String("username", username))
			}
			return
		}

		if msg.ID == nil {
			continue
		}

		if s.processMessage(ctx, conn, msg, &username) {
			return
		}
	}
}

func (s *Server) processMessage(ctx context.Context, conn net.Conn, msg *tcp.Message, username *string) bool {
	switch Method(msg.Method) {
	case MethodAuthorize:
		return s.handleAuthorizeMessage(ctx, conn, msg, username)
	case MethodSubmit:
		return s.handleSubmitMessage(ctx, conn, msg, *username)
	default:
		s.sendError(conn, *msg.ID, "unknown method")
		return false
	}
}

func (s *Server) handleAuthorizeMessage(ctx context.Context, conn net.Conn, msg *tcp.Message, username *string) bool {
	user, err := s.authorize(ctx, conn, msg)
	if err != nil {
		logger.Error("authorization failed", zap.Error(err))
		return true
	}
	*username = user
	return false
}

func (s *Server) handleSubmitMessage(ctx context.Context, conn net.Conn, msg *tcp.Message, username string) bool {
	if username == "" {
		s.sendError(conn, *msg.ID, "unauthorized")
		return true
	}
	s.submit(ctx, conn, msg, username)
	return false
}

func (*Server) sendError(conn net.Conn, id int64, errMsg string) {
	response := tcp.NewErrorResponse(id, errMsg)
	if err := tcp.WriteMessage(conn, response); err != nil {
		logger.Error("failed to send error response", zap.Error(err))
	}
}
