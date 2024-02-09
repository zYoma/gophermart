package main

import (
	"errors"

	"github.com/zYoma/gophermart/internal/app"
	"github.com/zYoma/gophermart/internal/config"
	"github.com/zYoma/gophermart/internal/logger"
)

func main() {
	// получаем конфигурацию
	cfg := config.GetConfig()

	// инициализируем логер
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		panic(err)
	}

	// инициализация приложения
	application, err := app.New(cfg)
	if err != nil {
		panic(err)
	}

	// запускаем приложение
	if err := application.Run(); err != nil {
		if errors.Is(err, app.ErrServerStoped) {
			logger.Log.Sugar().Infoln("stopping application")
			return
		}

		panic(err)
	}
}
