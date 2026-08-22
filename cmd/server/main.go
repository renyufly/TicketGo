package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"ticketgo/internal/app"
	"ticketgo/internal/config"
	logpkg "ticketgo/pkg/logger"
)

func main() {
	logger, err := logpkg.New()
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("invalid configuration", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, cfg, logger); err != nil {
		logger.Fatal("server stopped unexpectedly", zap.Error(err))
	}
}
