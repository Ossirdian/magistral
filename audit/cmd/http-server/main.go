package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/ossirdian/magistral/audit/internal/app"
	"github.com/ossirdian/magistral/audit/internal/config"
	"github.com/ossirdian/magistral/platform/pkg/closer"
	"github.com/ossirdian/magistral/platform/pkg/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.FromEnv()
	if err := logger.Init(cfg.LoggerLevel, cfg.LoggerAsJSON); err != nil {
		panic(err)
	}
	defer func() { _ = logger.Global().Sync() }()
	closer.SetLogger(logger.Global())

	if err := run(ctx, cfg); err != nil && !errors.Is(err, context.Canceled) {
		logger.Global().Error("audit service failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config.Config) error {
	application, err := app.New(ctx, cfg)
	if err != nil {
		return err
	}
	return application.Run(ctx)
}
