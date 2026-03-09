package config

import (
	"log"
	"os"
	"time"
)

type Config struct {
	ENV             string
	Addr            string
	AuthServiceHost string
	JwtSecret       []byte
	TokenTTL        time.Duration
}

func LoadConfig() Config {
	env := os.Getenv("ENVIRONMENT")
	isDev := env == "development"

	addr := os.Getenv("AUTH_SERVICE_ADDR")
	if addr == "" {
		if !isDev {
			log.Fatal("AUTH_SERVICE_ADDR is required")
		}
		addr = ":8080"
	}

	authServiceHost := os.Getenv("AUTH_SERVICE_HOST")
	if authServiceHost == "" {
		if !isDev {
			log.Fatal("AUTH_SERVICE_HOST is required")
		}
		authServiceHost = "http://localhost:8080"
	}

	secretKey := []byte(os.Getenv("JWT_SECRET"))
	if len(secretKey) == 0 {
		if !isDev {
			log.Fatal("JWT_SECRET is required")
		}
		secretKey = []byte("secret")
	}

	tokenTTLStr := os.Getenv("TOKEN_TTL")
	if tokenTTLStr == "" {
		if !isDev {
			log.Fatal("TOKEN_TTL is required")
		}
		tokenTTLStr = "1h"
	}

	tokenTTL, err := time.ParseDuration(tokenTTLStr)
	if err != nil {
		log.Fatal("TOKEN_TTL is invalid")
	}
	return Config{
		ENV:             env,
		Addr:            addr,
		AuthServiceHost: authServiceHost,
		JwtSecret:       secretKey,
		TokenTTL:        tokenTTL,
	}
}
