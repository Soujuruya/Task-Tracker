package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"task-tracker-1/internal/repository/user"
	"task-tracker-1/internal/service"
	"task-tracker-1/internal/transport"
	"task-tracker-1/internal/transport/handlers"
	"time"
)

type config struct {
	addr      string
	jwtSecret []byte
	tokenTTL  time.Duration
}

func loadConfig() config {
	addr := os.Getenv("AUTH_SERVICE_ADDR")
	if addr == "" {
		log.Fatal("AUTH_SERVICE_ADDR is required")
	}

	secretKey := []byte(os.Getenv("JWT_SECRET"))
	if len(secretKey) == 0 {
		log.Fatal("JWT_SECRET is required")
	}

	tokenTTLStr := os.Getenv("TOKEN_TTL")
	if tokenTTLStr == "" {
		log.Fatal("TOKEN_TTL is required")
	}

	tokenTTL, err := time.ParseDuration(tokenTTLStr)
	if err != nil {
		log.Fatal("TOKEN_TTL is invalid")
	}
	return config{
		addr:      addr,
		jwtSecret: secretKey,
		tokenTTL:  tokenTTL,
	}
}
func main() {

	// Парсим переменные окружения,в дальнейшем можно вынести работу с конфигурацией из main.
	cfg := loadConfig()

	userRepo := user.NewUserRepository()

	authService := service.NewAuthService(userRepo)

	authHandler := handlers.NewAuthHandler(authService, cfg.jwtSecret, cfg.tokenTTL)

	server := transport.NewServer(authHandler, cfg.addr)

	// Ловим сигналы SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("starting server at %s", cfg.addr)
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
