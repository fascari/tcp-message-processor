package application

import (
	"context"
	"net"

	"tcp-message-processor/common/pkg/logger"
	"tcp-message-processor/internal/config"
	"tcp-message-processor/internal/handler"
	"tcp-message-processor/internal/infra/consumer"
	"tcp-message-processor/internal/infra/publisher"
	"tcp-message-processor/internal/session"

	"go.uber.org/zap"
)

type Server struct {
	cfg       config.Config
	listener  net.Listener
	handler   *handler.Server
	consumer  *consumer.Consumer
	publisher *publisher.Publisher
	ctx       context.Context
	cancel    context.CancelFunc
}

func New(cfg config.Config) (Server, error) {
	statsStore, err := initDatabase(cfg.Database)
	if err != nil {
		return Server{}, err
	}

	pub, cons, err := initMessaging(cfg.RabbitMQ, statsStore)
	if err != nil {
		return Server{}, err
	}

	listener, err := initListener(cfg.Server)
	if err != nil {
		logger.Error("failed to initialize listener", zap.Error(err))
		cleanup(cons, pub)
		return Server{}, err
	}

	h := handler.New(session.NewStore(), pub, cfg.Server.BroadcastInterval)
	ctx, cancel := context.WithCancel(context.Background())

	return Server{
		cfg:       cfg,
		listener:  listener,
		handler:   h,
		consumer:  cons,
		publisher: pub,
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}
