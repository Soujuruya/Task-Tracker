package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"task-tracker-1/internal/pkg"
	"task-tracker-1/internal/pkg/ctxkeys"
)

func ValidateTokenMiddleware(validator AccessTokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			requestID := getRequestID(r)

			token, err := parseBearerToken(r)
			if err != nil {
				slog.Warn("authorization header is invalid", "request_id", requestID)
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			validatedTokenClaims, err := validator.ValidateAccessToken(r.Context(), token)
			if err != nil {
				slog.Warn("token validation failed", "request_id", requestID, "error", err)
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			if validatedTokenClaims.UserID == "" {
				slog.Warn("token validation returned empty user_id", "request_id", requestID)
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			slog.Debug("token validated successfully", "request_id", requestID, "user_id", validatedTokenClaims.UserID)
			ctx := context.WithValue(r.Context(), ctxkeys.UserIDKey, validatedTokenClaims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := getRequestID(r)
			key, err := makeRateLimitKey(r)
			if err != nil {
				slog.Error("failed to build rate limit key", "request_id", requestID, "error", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			isAllowed := limiter.AllowRequest(key)
			if !isAllowed {
				slog.Warn("rate limit exceeded", "request_id", requestID, "method", r.Method, "path", r.URL.Path)
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
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
