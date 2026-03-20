package config

import (
	"log"
	"os"
	"runtime"
	"strconv"
	"time"
)

type Config struct {
	ENV             string
	Addr            string
	AuthServiceHost string
	JwtSecret       []byte
	TokenTTL        time.Duration
	HashParameters
}

// HashParameters Отдельная структура под параметры
type HashParameters struct {
	Memory         uint32 //сколько памяти использует алгоритм (в коде килобайты)
	Iterations     uint32 // кол-во проходов по памяти
	Parallelism    uint8  // кол-во параллельных потоков внутри одного хеширования
	SaltLength     uint32 // длина случайной соли(случайный набор байт)
	KeyLength      uint32 // длина итогового хэша в байтах
	MaxConcurrency uint   // кол-во одновременных хэширований на сервере
}

func LoadConfig() Config {
	env := os.Getenv("ENVIRONMENT")
	isDev := env == "development"

	hashParameters := parseArgon2HashParameters(isDev)

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
		HashParameters:  hashParameters,
	}
}

// Отдельная функция для парсинга параметров хэширования
func parseArgon2HashParameters(isDev bool) HashParameters {
	memory := os.Getenv("MEMORY")
	if memory == "" {
		if !isDev {
			log.Fatal("MEMORY is required")
		}
		memory = "64"
	}
	iterations := os.Getenv("ITERATIONS")
	if iterations == "" {
		if !isDev {
			log.Fatal("ITERATIONS is required")
		}
		iterations = "3"
	}
	parallelism := os.Getenv("PARALLELISM")
	if parallelism == "" {
		if !isDev {
			log.Fatal("PARALLELISM is required")
		}
		parallelism = "2"
	}
	saltLength := os.Getenv("SALT_LENGTH")
	if saltLength == "" {
		if !isDev {
			log.Fatal("SALT_LENGTH is required")
		}
		saltLength = "16"
	}
	keyLength := os.Getenv("KEY_LENGTH")
	if keyLength == "" {
		if !isDev {
			log.Fatal("KEY_LENGTH is required")
		}
		keyLength = "32"
	}
	maxConcurrency := os.Getenv("MAX_CONCURRENCY")
	if maxConcurrency == "" {
		if !isDev {
			log.Fatal("MAX_CONCURRENCY is required")
		}
		maxConcurrency = strconv.Itoa(runtime.GOMAXPROCS(0))
	}
	return HashParameters{
		Memory:         parseUint32(memory) * 1024,
		Iterations:     parseUint32(iterations),
		Parallelism:    parseUint8(parallelism),
		SaltLength:     parseUint32(saltLength),
		KeyLength:      parseUint32(keyLength),
		MaxConcurrency: parseUint(maxConcurrency),
	}
}

func parseUint32(val string) uint32 {
	n, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		log.Fatalf("invalid value %s: %v", val, err)
	}
	return uint32(n)
}

func parseUint8(val string) uint8 {
	n, err := strconv.ParseUint(val, 10, 8)
	if err != nil {
		log.Fatalf("invalid value %s: %v", val, err)
	}
	return uint8(n)
}

func parseUint(val string) uint {
	n, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		log.Fatalf("invalid value %s: %v", val, err)
	}
	return uint(n)
}
