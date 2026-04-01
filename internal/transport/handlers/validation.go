package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"task-tracker-1/internal/domain"
	"task-tracker-1/internal/transport/dto"
)

// Добавлена новая функция обработки ошибок в TaskHandlers
func writeTaskError(w http.ResponseWriter, requestID string, err error) bool {
	if err == nil {
		return false
	}
	slog.Error("task error", "request_id", requestID, "error", err)
	switch {
	case errors.Is(err, domain.ErrTaskNotFound):
		http.Error(w, "task not found", http.StatusNotFound)
	case errors.Is(err, domain.ErrEmptyTaskTitle):
		http.Error(w, "empty task title", http.StatusBadRequest)
	case errors.Is(err, domain.ErrTaskAlreadyDone):
		http.Error(w, "task already done", http.StatusBadRequest)
	case errors.Is(err, domain.ErrInvalidTaskStatus):
		http.Error(w, "invalid task status", http.StatusBadRequest)
	case errors.Is(err, domain.ErrInvalidTransition):
		http.Error(w, "invalid task transition", http.StatusConflict)
	case errors.Is(err, domain.ErrInvalidUserID):
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	case errors.Is(err, domain.ErrMissingPathID):
		http.Error(w, "id is required", http.StatusBadRequest)
	case errors.Is(err, domain.ErrInvalidQueryParam):
		http.Error(w, "invalid query param", http.StatusBadRequest)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
	return true
}

// Добавлена фукнция обработки ошибок в AuthHandlers
func writeAuthError(w http.ResponseWriter, requestID string, err error) bool {
	if err == nil {
		return false
	}
	slog.Error("auth error", "request_id", requestID, "error", err)
	switch {
	case errors.Is(err, domain.ErrUserAlreadyExists):
		http.Error(w, "user already exists", http.StatusConflict)
	case errors.Is(err, domain.ErrUserNameRequired):
		http.Error(w, "username is required", http.StatusBadRequest)
	case errors.Is(err, domain.ErrPasswordRequired):
		http.Error(w, "password is required", http.StatusBadRequest)
	case errors.Is(err, domain.ErrCredentialsRequired):
		http.Error(w, "credentials are required", http.StatusBadRequest)
	case errors.Is(err, domain.ErrInvalidCredentials):
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
	case errors.Is(err, domain.ErrInvalidUserID):
		http.Error(w, "invalid userID", http.StatusUnauthorized)
	case errors.Is(err, domain.ErrTokenInvalid):
		http.Error(w, "token invalid", http.StatusUnauthorized)
	case errors.Is(err, domain.ErrTokenExpired):
		http.Error(w, "token expired", http.StatusUnauthorized)
	case errors.Is(err, domain.ErrMissingToken):
		http.Error(w, "missing token", http.StatusUnauthorized)
	case errors.Is(err, domain.ErrGenerateToken):
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
	case errors.Is(err, domain.ErrRefreshTokenInvalid):
		http.Error(w, "invalid refresh token", http.StatusBadRequest)
	case errors.Is(err, domain.ErrRefreshTokenExpired):
		http.Error(w, "refresh token expired", http.StatusUnauthorized)
	case errors.Is(err, domain.ErrRefreshTokenNotFound):
		http.Error(w, "refresh token not found", http.StatusUnauthorized)
	case errors.Is(err, domain.ErrRefreshTokenGenerate):
		http.Error(w, "failed to generate refresh token", http.StatusInternalServerError)
	case errors.Is(err, domain.ErrRefreshTokenAlreadyExists):
		http.Error(w, "refresh token already exists", http.StatusConflict)
	case errors.Is(err, domain.ErrActiveRefreshTokenAlreadyExists):
		http.Error(w, "active refresh token already exists", http.StatusConflict)
	case errors.Is(err, domain.ErrRefreshTokenRevoked):
		http.Error(w, "refresh token revoked", http.StatusUnauthorized)
	case errors.Is(err, domain.ErrRefreshTokenRotated):
		http.Error(w, "refresh token rotated", http.StatusUnauthorized)
	case errors.Is(err, domain.ErrInvalidRefreshTokenState):
		http.Error(w, "invalid refresh token state", http.StatusInternalServerError)
	default:
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

// Фукнция валидации данных пагинации
func isInvalidPagination(page, pageSize int) bool {
	if page < 1 {
		return true
	}
	if pageSize > 100 || pageSize < 1 {
		return true
	}
	return false
}

func validateRefreshRequest(req dto.RefreshRequest) error {
	if req.RefreshToken == "" {
		return domain.ErrRefreshTokenInvalid
	}
	return nil
}
