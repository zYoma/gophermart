package main

import (
	"context"
	"errors"
	"os/signal"
	"syscall"

	"github.com/zYoma/gophermart/internal/app"
	"github.com/zYoma/gophermart/internal/config"
	"github.com/zYoma/gophermart/internal/logger"
)

func main() {
	// получаем конфигурацию
	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	// инициализируем логер
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// инициализация приложения
	application, err := app.New(ctx, cfg)
	if err != nil {
		panic(err)
	}

	// запускаем приложение
	if err := application.Run(ctx); err != nil {
		if errors.Is(err, app.ErrServerStoped) {
			logger.Log.Sugar().Infoln("stopping application")
			return
		}

		panic(err)
	}
}
