package application

import (
	"context"

	"tcp-message-processor/common/pkg/logger"

	"go.uber.org/zap"
)

func (a Server) acceptConnections(ctx context.Context) {
	for {
		if a.shouldStop() {
			return
		}

		conn, err := a.listener.Accept()
		if err != nil {
			if a.shouldStop() {
				return
			}
			logger.Error("failed to accept connection", zap.Error(err))
			continue
		}

		logger.Info("new connection accepted", zap.String("remote_addr", conn.RemoteAddr().String()))
		go a.handler.Handle(ctx, conn)
	}
}

func (a Server) shouldStop() bool {
	select {
	case <-a.ctx.Done():
		return true
	default:
		return false
	}
}
