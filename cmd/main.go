package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"task-tracker-1/internal/config"
	loclog "task-tracker-1/internal/pkg/logger"
	"task-tracker-1/internal/repository/task"
	"task-tracker-1/internal/repository/user"
	"task-tracker-1/internal/service"
	"task-tracker-1/internal/transport"
	"task-tracker-1/internal/transport/handlers"
	"time"
)

func main() {
	logger := loclog.Init()
	slog.SetDefault(logger)

	// Парсим переменные окружения из .env
	cfg := config.LoadConfig()
	slog.Info("Loading config...", "ENVIRONMENT", cfg.ENV)

	userRepo := user.NewUserRepository()
	authService := service.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService, cfg.JwtSecret, cfg.TokenTTL)

	taskRepo := task.NewTaskRepository()
	taskService := service.NewTaskService(taskRepo)
	taskHandler := handlers.NewTaskHandler(taskService)

	server := transport.NewServer(authHandler, taskHandler, cfg.Addr, cfg.AuthServiceHost)

	// Ловим сигналы SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("starting server at %s", cfg.Addr)
		if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) { // игнорируем ErrServerClosed ошибку для чистого завершения
			log.Fatal("failed to start server:", err)
		}
	}()

	// Ждём сигнал завершения
	<-ctx.Done()
	log.Println("shutting down server...")

	// Контекст с таймаутом для завершения текущих запросов
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Сервер перестаёт принимать запросы,дожидается завершения текущих запросов и выключается
	// Контекст нужен, чтобы сервер бесконечно не ждал завершения запросов, 10 секунд и принудительно завершаем.
	if err := server.GracefulShutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("failed to shutdown server:", err)
	}

	log.Println("shutting down gracefully")
}
