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
	JwtSecret       []byte
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration

	HashParameters
	RateLimiterConfig
	StorageConfig
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

type RateLimiterConfig struct {
	MaxRequests int
	WindowSize  time.Duration
}

type StorageConfig struct {
	StorageType string
	PostgresDSN string
}

func LoadConfig() Config {
	env := os.Getenv("ENVIRONMENT")
	isDev := env == "development"

	hashParameters := parseArgon2HashParameters(isDev)
	rateLimiterConfig := parseRateLimiterConfig(isDev)

	addr := os.Getenv("AUTH_SERVICE_ADDR")
	if addr == "" {
		if !isDev {
			log.Fatal("AUTH_SERVICE_ADDR is required")
		}
		addr = ":8080"
	}

	secretKey := []byte(os.Getenv("JWT_SECRET"))
	if len(secretKey) == 0 {
		if !isDev {
			log.Fatal("JWT_SECRET is required")
		}
		secretKey = []byte("secret")
	}

	accessTokenTTLStr := os.Getenv("ACCESS_TOKEN_TTL")
	if accessTokenTTLStr == "" {
		if !isDev {
			log.Fatal("ACCESS_TOKEN_TTL is required")
		}
		accessTokenTTLStr = "1h"
	}
	refreshTokenTTLStr := os.Getenv("REFRESH_TOKEN_TTL")
	if refreshTokenTTLStr == "" {
		if !isDev {
			log.Fatal("REFRESH_TOKEN_TTL is required")
		}
		refreshTokenTTLStr = "168h"
	}

	accessTokenTTL, err := time.ParseDuration(accessTokenTTLStr)
	if err != nil {
		log.Fatal("ACCESS_TOKEN_TTL is invalid")
	}
	refreshTokenTTL, err := time.ParseDuration(refreshTokenTTLStr)
	if err != nil {
		log.Fatal("REFRESH_TOKEN_TTL is invalid")
	}

	storageConfig := parseStorageConfig(isDev)

	return Config{
		ENV:               env,
		Addr:              addr,
		JwtSecret:         secretKey,
		AccessTokenTTL:    accessTokenTTL,
		RefreshTokenTTL:   refreshTokenTTL,
		HashParameters:    hashParameters,
		RateLimiterConfig: rateLimiterConfig,
		StorageConfig:     storageConfig,
	}
}

// Функция парсинга параметров хранения данных
func parseStorageConfig(isDev bool) StorageConfig {
	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "" {
		if !isDev {
			log.Fatal("STORAGE_TYPE is required")
		}
		storageType = "memory"
	}

	switch storageType {
	case "memory", "postgres":
	default:
		log.Fatal("STORAGE_TYPE must be either memory or postgres")
	}

	postgresDSN := ""
	if storageType == "postgres" {
		postgresDSN = os.Getenv("POSTGRES_DSN")
		if postgresDSN == "" {
			if !isDev {
				log.Fatal("POSTGRES_DSN is required")
			}
			postgresDSN = "postgres://user:password@localhost:5432/task_tracker?sslmode=disable"
		}
	}

	return StorageConfig{
		StorageType: storageType,
		PostgresDSN: postgresDSN,
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

func parseRateLimiterConfig(isDev bool) RateLimiterConfig {
	maxRequestsStr := os.Getenv("RATE_LIMIT_MAX_REQUESTS")
	if maxRequestsStr == "" {
		if !isDev {
			log.Fatal("RATE_LIMIT_MAX_REQUESTS is required")
		}
		maxRequestsStr = "10"
	}
	maxRequests, err := strconv.Atoi(maxRequestsStr)
	if err != nil {
		log.Fatal("RATE_LIMIT_MAX_REQUESTS is invalid")
	}

	windowSizeStr := os.Getenv("RATE_LIMIT_WINDOW_SIZE")
	if windowSizeStr == "" {
		if !isDev {
			log.Fatal("RATE_LIMIT_WINDOW_SIZE is required")
		}
		windowSizeStr = "1m"
	}
	windowSize, err := time.ParseDuration(windowSizeStr)
	if err != nil {
		log.Fatal("RATE_LIMIT_WINDOW_SIZE is invalid")
	}

	return RateLimiterConfig{
		MaxRequests: maxRequests,
		WindowSize:  windowSize,
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
