package main

import (
    "context"
    "cert-secret-applier/config"
    "cert-secret-applier/logger"
    "cert-secret-applier/app"
	"go.uber.org/zap"
)

func Run(ctx context.Context) error {
    if err := config.LoadConfig("config", "config"); err != nil {
		return err
	}

    cfg := config.Global()

    if err := logger.Init(cfg.Logger.Level); err != nil {
        logger.ErrorCtx(ctx, "Logger init", zap.Error(err))
		return err
	}
	logger.DebugCtx(ctx, "configuration", zap.Any("config", cfg))

    if err := app.Run(cfg); err != nil {
		return err
	}

    return nil
}
