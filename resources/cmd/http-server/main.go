package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/ossirdian/magistral/platform/pkg/closer"
	"github.com/ossirdian/magistral/platform/pkg/logger"
	"github.com/ossirdian/magistral/resources/internal/app"
	cfgenv "github.com/ossirdian/magistral/resources/internal/config/env"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := cfgenv.New()
	if err != nil {
		panic(err)
	}

	if err := logger.Init(cfg.Logger().Level(), cfg.Logger().AsJSON()); err != nil {
		panic(err)
	}
	defer func() { _ = logger.Global().Sync() }()

	closer.SetLogger(logger.Global())

	if err := run(ctx, cfg); err != nil && !errors.Is(err, context.Canceled) {
		logger.Global().Error("resources service failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg *cfgenv.Config) error {
	application, err := app.New(ctx, cfg)
	if err != nil {
		return err
	}
	return application.Run(ctx)
}
