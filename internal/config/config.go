package config

import (
	"log"
	"os"
	"time"
)

type Config struct {
	Addr      string
	JwtSecret []byte
	TokenTTL  time.Duration
}

func LoadConfig() Config {
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
	return Config{
		Addr:      addr,
		JwtSecret: secretKey,
		TokenTTL:  tokenTTL,
	}
}
