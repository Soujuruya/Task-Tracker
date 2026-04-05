package middleware

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"task-tracker-1/internal/pkg/ctxkeys"
)

func parseBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	isBearer := strings.HasPrefix(authHeader, "Bearer ")
	if !isBearer {
		return "", errors.New("invalid bearer token")
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", errors.New("invalid bearer token")
	}
	return token, nil
}

func makeRateLimitKey(r *http.Request) (string, error) {
	method := r.Method
	path := r.URL.Path

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", fmt.Errorf("invalid remote address: %w", err)
	}

	if ip == "" || method == "" || path == "" {
		return "", errors.New("invalid rate limit key")
	}

	key := strings.Join([]string{ip, method, path}, "|")

	return key, nil
}

func getRequestID(r *http.Request) string {
	id, _ := r.Context().Value(ctxkeys.RequestIDKey).(string)
	return id
}
