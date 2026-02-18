package main

import (
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"auth-service/internal/transport"
	"log"
	"time"
)

func main() {
	// Временное решение для быстроты реализации и облегчения запуска
	addr := ":8080"
	secretKey := []byte("secretkeyjwt") // для JWT
	tokenTTL := 1 * time.Hour           // время жизни токена

	userRepo := repository.NewUserRepository()

	authService := service.NewAuthService(userRepo)

	authHandler := transport.NewAuthHandler(authService, secretKey, tokenTTL)

	server := transport.NewServer(authHandler, addr)

	log.Printf("Starting server on %s\n", addr)
	if err := server.Start(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
