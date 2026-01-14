package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"tcp-message-processor/common/pkg/logger"
	"tcp-message-processor/internal/config"
	"tcp-message-processor/internal/infra/application"

	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		logger.Error("server error", zap.Error(err))
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if err := logger.Init(); err != nil {
		return err
	}

	logger.Info("starting tcp message processor server", zap.String("version", "1.0.0"))

	srv, err := application.New(cfg)
	if err != nil {
		return err
	}

	if err := srv.Start(ctx); err != nil {
		return err
	}

	waitForShutdown()

	logger.Info("shutting down server...")
	srv.Stop()
	logger.Info("server stopped")

	return nil
}

func waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
