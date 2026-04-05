package middleware

import (
	"errors"
	"sync"
	"time"
)

type RateLimitConfig struct {
	MaxRequests int
	WindowSize  time.Duration
}

type RateLimiter struct {
	mu          sync.Mutex
	limitConfig RateLimitConfig
	requestLogs map[string][]time.Time
}

func NewRateLimiter(limitConfig RateLimitConfig) (*RateLimiter, error) {
	if limitConfig.MaxRequests <= 0 || limitConfig.WindowSize <= 0 {
		return nil, errors.New("invalid rate limit configuration")
	}
	return &RateLimiter{
		limitConfig: limitConfig,
		requestLogs: make(map[string][]time.Time),
	}, nil
}

// AllowRequest Алгоритм Sliding Window Log
func (rl *RateLimiter) AllowRequest(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	currentTime := time.Now()
	// Получаем истории запросов по ключу
	requests, exists := rl.requestLogs[key]
	if !exists {
		requests = []time.Time{}
	}
	// Определяем границу окна
	cutoff := currentTime.Add(-rl.limitConfig.WindowSize)
	var newLog []time.Time

	// Оставляем запросы только внутри окна
	for _, t := range requests {
		if t.After(cutoff) {
			newLog = append(newLog, t)
		}
	}
	// Если лимит не достигнут, сохраняем запрос и разрешаем его
	if len(newLog) < rl.limitConfig.MaxRequests {
		newLog = append(newLog, currentTime)
		rl.requestLogs[key] = newLog
		return true
	}
	// Иначе отклоняем запрос
	rl.requestLogs[key] = newLog
	return false
}
