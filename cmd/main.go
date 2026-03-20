package main

import (
	"log"
	"log/slog"
	"task-tracker-1/internal/app"
	"task-tracker-1/internal/config"
	loclog "task-tracker-1/internal/pkg/logger"
)

func main() {
	// Парсим переменные окружения из .env
	cfg := config.LoadConfig()

	logger := loclog.Init(cfg.ENV)
	slog.SetDefault(logger)

	slog.Info("Loading config...", "ENVIRONMENT", cfg.ENV)
	slog.Info("Server run", "addr", cfg.Addr)
	// Собираем всё вместе и запускаем
	app := app.NewApp(&cfg)
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
