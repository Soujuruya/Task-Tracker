package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"task-tracker-1/internal/transport/dto"
)

type contextKey string

const UserIDKey contextKey = "userID"

func ValidateTokenMiddleware(authServiceURl string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := r.Header.Get("Authorization")
			token := strings.TrimPrefix(tokenString, "Bearer ")
			if token == "" {
				http.Error(w, "missing token", http.StatusUnauthorized)
				return
			}

			req, err := http.NewRequest(http.MethodGet, authServiceURl+"/validate", nil)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			req.Header.Set("Authorization", "Bearer "+token)

			client := &http.Client{}
			resp, err := client.Do(req)
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
			ctx := context.WithValue(r.Context(), UserIDKey, validateResponse.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
