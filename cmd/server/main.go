package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"tcp-message-processor/internal/config"
	pkgapp "tcp-message-processor/internal/infra/app"
	"tcp-message-processor/pkg/logger"

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

	app, err := pkgapp.New(cfg)
	if err != nil {
		return err
	}

	if err := app.Start(ctx); err != nil {
		return err
	}

	waitForShutdown()

	logger.Info("shutting down server...")
	app.Stop()
	logger.Info("server stopped")

	return nil
}

func waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
