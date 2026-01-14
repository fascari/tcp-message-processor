package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"tcp-message-processor-client/internal/app"
	"tcp-message-processor-client/internal/config"
	"tcp-message-processor/common/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		logger.Error("client error", zap.Error(err))
		os.Exit(1)
	}
}

func run() error {
	if err := logger.Init(); err != nil {
		return err
	}
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger.Info("starting tcp client", zap.String("username", cfg.Client.Username))

	client, err := app.New(cfg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		waitForShutdown()
		logger.Info("shutting down client...")
		cancel()
	}()

	return client.Run(ctx)
}

func waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
