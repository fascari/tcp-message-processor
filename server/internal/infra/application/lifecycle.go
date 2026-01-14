package application

import (
	"context"

	"tcp-message-processor/common/pkg/logger"

	"go.uber.org/zap"
)

func (a Server) Start(ctx context.Context) error {
	if err := a.consumer.Start(ctx); err != nil {
		return err
	}

	a.handler.Start()
	go a.acceptConnections(ctx)

	return nil
}

func (a Server) Stop() {
	a.cancel()
	a.handler.Stop()

	a.closeListener()
	a.stopConsumer()
	a.closePublisher()

	logger.Sync()
}

func (a Server) closeListener() {
	if a.listener != nil {
		if err := a.listener.Close(); err != nil {
			logger.Error("listener close error", zap.Error(err))
		}
	}
}

func (a Server) stopConsumer() {
	if a.consumer != nil {
		if err := a.consumer.Stop(); err != nil {
			logger.Error("consumer stop error", zap.Error(err))
		}
	}
}

func (a Server) closePublisher() {
	if a.publisher != nil {
		if err := a.publisher.Close(); err != nil {
			logger.Error("publisher close error", zap.Error(err))
		}
	}
}
