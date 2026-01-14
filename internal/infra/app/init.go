package app

import (
	"fmt"
	"net"

	"tcp-message-processor/internal/config"
	"tcp-message-processor/internal/infra/consumer"
	"tcp-message-processor/internal/infra/database"
	"tcp-message-processor/internal/infra/publisher"
	"tcp-message-processor/internal/stats"
	"tcp-message-processor/pkg/closer"
	"tcp-message-processor/pkg/logger"

	"go.uber.org/zap"
)

func initDatabase(cfg config.DatabaseConfig) (stats.Store, error) {
	db, err := database.New(cfg)
	if err != nil {
		return stats.Store{}, fmt.Errorf("failed to initialize database: %w", err)
	}
	return stats.NewStore(db), nil
}

func initMessaging(cfg config.RabbitMQConfig, statsStore stats.Store) (*publisher.Publisher, *consumer.Consumer, error) {
	pub, err := publisher.New(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize publisher: %w", err)
	}

	cons, err := consumer.New(cfg, statsStore)
	if err != nil {
		closer.Close(pub, "failed to close publisher during cleanup")
		return nil, nil, fmt.Errorf("failed to initialize consumer: %w", err)
	}

	return pub, cons, nil
}

func initListener(cfg config.ServerConfig) (net.Listener, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	logger.Info("tcp server listening", zap.String("address", addr))
	return listener, nil
}

func cleanup(cons *consumer.Consumer, pub *publisher.Publisher) {
	if cons != nil {
		if err := cons.Stop(); err != nil {
			logger.Error("consumer stop error", zap.Error(err))
		}
	}
	closer.Close(pub, "publisher close error")
}
