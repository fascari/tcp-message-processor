package app

import (
	"context"

	"tcp-message-processor/pkg/logger"

	"go.uber.org/zap"
)

func (a *Application) Start(ctx context.Context) error {
	if err := a.consumer.Start(ctx); err != nil {
		return err
	}

	a.server.Start(ctx)
	go a.acceptConnections(ctx)

	return nil
}

func (a *Application) Stop() {
	a.cancel()
	a.server.Stop()

	a.closeListener()
	a.stopConsumer()
	a.closePublisher()

	logger.Sync()
}

func (a *Application) closeListener() {
	if a.listener != nil {
		if err := a.listener.Close(); err != nil {
			logger.Error("listener close error", zap.Error(err))
		}
	}
}

func (a *Application) stopConsumer() {
	if a.consumer != nil {
		if err := a.consumer.Stop(); err != nil {
			logger.Error("consumer stop error", zap.Error(err))
		}
	}
}

func (a *Application) closePublisher() {
	if a.publisher != nil {
		if err := a.publisher.Close(); err != nil {
			logger.Error("publisher close error", zap.Error(err))
		}
	}
}
