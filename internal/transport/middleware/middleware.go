package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"task-tracker-1/internal/pkg"
	"task-tracker-1/internal/pkg/ctxkeys"
	"task-tracker-1/internal/transport/dto"
	"time"
)

var httpClient = &http.Client{Timeout: 3 * time.Second}

func ValidateTokenMiddleware(authServiceURl string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			tokenString := r.Header.Get("Authorization")
			token := strings.TrimPrefix(tokenString, "Bearer ")
			if token == "" {
				http.Error(w, "missing token", http.StatusUnauthorized)
				return
			}

			req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, authServiceURl+"/validate", nil)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := httpClient.Do(req)
			if err != nil {
				http.Error(w, "auth service unavailable", http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				http.Error(w, "validation failed", http.StatusUnauthorized)
				return
			}

			var validateResponse dto.ValidateTokenResponse

			if err := json.NewDecoder(resp.Body).Decode(&validateResponse); err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			if validateResponse.UserID == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ctxkeys.UserIDKey, validateResponse.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID, err := pkg.GenerateID()
		if err != nil {
			requestID = "unknown"
		}
		ctx := context.WithValue(r.Context(), ctxkeys.RequestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID, _ := r.Context().Value(ctxkeys.RequestIDKey).(string)

		slog.Info("incoming request",
			"request_id", requestID,
			"method", r.Method,
			"url", r.URL.Path,
		)

		next.ServeHTTP(w, r)
	})
}
