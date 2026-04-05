package middleware

import (
	"errors"
	"net/http"
	"strings"
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
