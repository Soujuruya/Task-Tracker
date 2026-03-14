package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"task-tracker-1/internal/domain"
	"task-tracker-1/internal/transport/dto"
	"task-tracker-1/internal/transport/middleware"
)

// Добавлена новая функция для парсинга токена из заголовка Authorization
func parseBearerToken(r *http.Request) (string, error) {
	tokenString := r.Header.Get("Authorization")
	token := strings.TrimPrefix(tokenString, "Bearer ")
	if token == "" {
		return "", domain.ErrMissingToken
	}
	return token, nil
}

// Добавлена новая фукнция извлечения из контекста UserID
func getUserID(r *http.Request) (string, error) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		slog.Error("unauthorized: missing or invalid userID")
		return "", domain.ErrInvalidUserID
	}
	return userID, nil
}

// Добавлена фукнция извлечения ID из path
func getPathID(r *http.Request) (string, error) {
	id := r.PathValue("id")
	if id == "" {
		return "", domain.ErrMissingPathID
	}
	return id, nil
}

// Добавлена новая функция обработки ошибок в TaskHandlers
func writeTaskError(w http.ResponseWriter, object string, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, domain.ErrTaskNotFound):
		slog.Error(object, "error", err)
		http.Error(w, "task not found", http.StatusNotFound)
	case errors.Is(err, domain.ErrEmptyTaskTitle):
		slog.Error(object, "error", err)
		http.Error(w, "empty task title", http.StatusBadRequest)
	case errors.Is(err, domain.ErrTaskAlreadyDone):
		slog.Error(object, "error", err)
		http.Error(w, "task already done", http.StatusBadRequest)
	case errors.Is(err, domain.ErrInvalidTaskStatus):
		slog.Error(object, "error", err)
		http.Error(w, "invalid task status", http.StatusBadRequest)
	case errors.Is(err, domain.ErrInvalidTransition):
		slog.Error(object, "error", err)
		http.Error(w, "invalid task transition", http.StatusConflict)
	case errors.Is(err, domain.ErrInvalidUserID):
		slog.Error(object, "error", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	case errors.Is(err, domain.ErrMissingPathID):
		slog.Error(object, "error", err)
		http.Error(w, "id is required", http.StatusBadRequest)
	default:
		slog.Error(object, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
	return true
}

// Добавлена фукнция обработки ошибок в AuthHandlers
func writeAuthError(w http.ResponseWriter, object string, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, domain.ErrUserAlreadyExists):
		slog.Error(object, "error", err)
		http.Error(w, "user already exists", http.StatusConflict)
	case errors.Is(err, domain.ErrUserNameRequired):
		slog.Error(object, "error", err)
		http.Error(w, "username is required", http.StatusBadRequest)
	case errors.Is(err, domain.ErrPasswordRequired):
		slog.Error(object, "error", err)
		http.Error(w, "password is required", http.StatusBadRequest)
	case errors.Is(err, domain.ErrCredentialsRequired):
		slog.Error(object, "error", err)
		http.Error(w, "credentials are required", http.StatusBadRequest)
	case errors.Is(err, domain.ErrInvalidCredentials):
		slog.Error(object, "error", err)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
	case errors.Is(err, domain.ErrInvalidUserID):
		slog.Error(object, "error", err)
		http.Error(w, "invalid userID", http.StatusUnauthorized)
	case errors.Is(err, domain.ErrTokenInvalid):
		slog.Error(object, "error", err)
		http.Error(w, "token invalid", http.StatusUnauthorized)
	case errors.Is(err, domain.ErrTokenExpired):
		slog.Error(object, "error", err)
		http.Error(w, "token expired", http.StatusUnauthorized)
	case errors.Is(err, domain.ErrMissingToken):
		slog.Error(object, "error", err)
		http.Error(w, "missing token", http.StatusUnauthorized)
	case errors.Is(err, domain.ErrGenerateToken):
		slog.Error(object, "error", err)
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
	default:
		slog.Error(object, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
	return true
}

// Валидация регистрационных данных
func validateRegisterRequest(req dto.RegisterRequest) error {
	if req.Username == "" {
		return domain.ErrUserNameRequired
	}
	if req.Password == "" {
		return domain.ErrPasswordRequired
	}
	return nil
}

// Валидация данных аунтефикации
func validateLoginRequest(req dto.LoginRequest) error {
	if req.Username == "" || req.Password == "" {
		return domain.ErrCredentialsRequired
	}
	return nil
}

// Фукнция для декодирования JSON
func decodeJSON(w http.ResponseWriter, r *http.Request, to any) bool {
	if err := json.NewDecoder(r.Body).Decode(to); err != nil {
		slog.Error("invalid request body", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return false
	}
	return true
}

// Функция для кодирования в JSON
func encodeJSON(w http.ResponseWriter, status int, data any) bool {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("encode response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return false
	}
	return true
}
