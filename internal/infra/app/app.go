package app

import (
	"context"
	"net"

	"tcp-message-processor/internal/config"
	"tcp-message-processor/internal/infra/consumer"
	"tcp-message-processor/internal/infra/publisher"
	"tcp-message-processor/internal/server"
	"tcp-message-processor/internal/session"
	"tcp-message-processor/pkg/logger"

	"go.uber.org/zap"
)

type Application struct {
	cfg       config.Config
	listener  net.Listener
	server    server.Server
	consumer  *consumer.Consumer
	publisher *publisher.Publisher
	ctx       context.Context
	cancel    context.CancelFunc
}

func New(cfg config.Config) (Application, error) {
	statsStore, err := initDatabase(cfg.Database)
	if err != nil {
		return Application{}, err
	}

	pub, cons, err := initMessaging(cfg.RabbitMQ, statsStore)
	if err != nil {
		return Application{}, err
	}

	listener, err := initListener(cfg.Server)
	if err != nil {
		logger.Error("failed to initialize listener", zap.Error(err))
		cleanup(cons, pub)
		return Application{}, err
	}

	tcpServer := server.New(session.NewStore(), pub, cfg.Server.BroadcastInterval)
	ctx, cancel := context.WithCancel(context.Background())

	return Application{
		cfg:       cfg,
		listener:  listener,
		server:    tcpServer,
		consumer:  cons,
		publisher: pub,
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}
