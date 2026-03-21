package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"task-tracker-1/internal/config"
	"task-tracker-1/internal/pkg/hasher"
	"task-tracker-1/internal/repository/auditlog"
	"task-tracker-1/internal/repository/task"
	"task-tracker-1/internal/repository/user"
	"task-tracker-1/internal/service"
	"task-tracker-1/internal/transport"
	"task-tracker-1/internal/transport/handlers"
	"time"
)

// Вынес сбор всех зависимостей из main, так как уже main сильно разросся
type App struct {
	server *transport.Server
}

func NewApp(cfg *config.Config) *App {
	// Инициализируем репозитории
	userRepo := user.NewUserRepository()
	taskRepo := task.NewTaskRepository()
	auditLogRepo := auditlog.NewAuditLogRepository()
	// Инициализируем хэшер
	argon2Hashes := hasher.NewArgon2Hasher(hasher.Argon2Params{
		Memory:         cfg.Memory,
		Iterations:     cfg.Iterations,
		Parallelism:    cfg.Parallelism,
		SaltLength:     cfg.SaltLength,
		KeyLength:      cfg.KeyLength,
		MaxConcurrency: cfg.MaxConcurrency,
	})
	// Инициализируем сервисы
	authService := service.NewAuthService(userRepo, argon2Hashes)
	taskService := service.NewTaskService(taskRepo, auditLogRepo)
	tokenService := service.NewTokenService(cfg.JwtSecret, cfg.TokenTTL)
	// Инициализируем хендлеры
	authHandler := handlers.NewAuthHandler(authService, tokenService)
	taskHandler := handlers.NewTaskHandler(taskService)
	// Инициализируем http-сервер
	server := transport.NewServer(authHandler, taskHandler, cfg.Addr, cfg.AuthServiceHost)

	return &App{server: server}
}

func (app *App) Start() error {
	// Ловим сигналы SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		if err := app.server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) { // игнорируем ErrServerClosed ошибку для чистого завершения
			errCh <- err
		}
	}()

	// Ждём сигнал завершения
	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
	}

	log.Println("shutting down server...")

	// Контекст с таймаутом для завершения текущих запросов
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Сервер перестаёт принимать запросы,дожидается завершения текущих запросов и выключается
	// Контекст нужен, чтобы сервер бесконечно не ждал завершения запросов, 10 секунд и принудительно завершаем.
	if err := app.server.GracefulShutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	log.Println("shutting down gracefully")

	return nil
}
