package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"task-tracker-1/internal/domain"
	"task-tracker-1/internal/service"
	"task-tracker-1/internal/transport/dto"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	AuthService    *service.AuthService
	JWTSecretKey   []byte
	AccessTokenTTL time.Duration
}

func NewAuthHandler(authService *service.AuthService, secretKey []byte, tokenTTL time.Duration) *AuthHandler {
	return &AuthHandler{
		AuthService:    authService,
		JWTSecretKey:   secretKey,
		AccessTokenTTL: tokenTTL,
	}
}

// ValidateToken Валидация токена (внутренний эндпоинт)
func (h *AuthHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		slog.Warn("AuthService.Handlers.ValidateToken: method not allowed", "method", r.Method)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokenString := r.Header.Get("Authorization")
	token := strings.TrimPrefix(tokenString, "Bearer ")
	if token == "" {
		slog.Error("AuthService.Handlers.ValidateToken: token required")
		http.Error(w, "token required", http.StatusUnauthorized)
		return
	}

	slog.Info("AuthService.Handlers.ValidateToken: token received")

	var accessClaims struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		jwt.RegisteredClaims
	}

	_, err := jwt.ParseWithClaims(token, &accessClaims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			slog.Info("AuthService.Handlers.ValidateToken: unexpected signing method")
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return h.JWTSecretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			slog.Error("AuthService.Handlers.ValidateToken", "error", err)
			http.Error(w, "token expired", http.StatusUnauthorized)
			return
		}
		slog.Error("AuthService.Handlers.ValidateToken: invalid token", "error", err)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	resp := dto.ValidateTokenResponse{
		Valid:  true,
		UserID: accessClaims.UserID,
	}

	slog.Info("AuthService.Handlers.ValidateToken: token valid", "user_id", accessClaims.UserID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("AuthService.Handlers.ValidateToken: encode response", "error", err)
		http.Error(w, "failed to write response", http.StatusInternalServerError)
		return
	}
}

// Register Регистрация пользователя
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("AuthService.Handlers.Register: invalid request body", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	slog.Info("AuthService.Handlers.Register", "username", req.Username)

	if req.Username == "" {
		slog.Error("AuthService.Handlers.Register: username required")
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		slog.Error("AuthService.Handlers.Register: password required")
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}

	userID, err := h.AuthService.Register(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			slog.Error("AuthService.Handlers.Register", "error", err)
			http.Error(w, "user already exists", http.StatusConflict)
			return
		}
		slog.Error("AuthService.Handlers.Register", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	slog.Info("AuthService.Handlers.Register: registration successful", "user_id", userID)
	resp := struct {
		UserID string `json:"user_id"`
	}{
		UserID: userID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("AuthService.Handlers.Register: encode response", "error", err)
		http.Error(w, "failed to write response", http.StatusInternalServerError)
		return
	}
}

// Login Аутентификация и получение токена
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("AuthService.Handlers.Login: invalid request body", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	slog.Info("AuthService.Handlers.Login: login attempt", "username", req.Username)

	if req.Username == "" || req.Password == "" {
		slog.Error("AuthService.Handlers.Login: username or password is required")
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	userID, err := h.AuthService.Login(req.Username, req.Password)

	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			slog.Error("AuthService.Handlers.Login", "error", err)
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		slog.Error("AuthService.Handlers.Login: login failed", "username", req.Username, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	accessClaims := jwt.MapClaims{
		"user_id":  userID,
		"username": req.Username,
		"exp":      time.Now().Add(h.AccessTokenTTL).Unix(),
		"iat":      time.Now().Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)

	signedToken, err := accessToken.SignedString(h.JWTSecretKey)
	if err != nil {
		slog.Error("AuthService.Handlers.Login: failed to generate token", "username", req.Username, "error", err)
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	slog.Info("AuthService.Handlers.Login: login successful", "username", req.Username, "user_id", userID)
	resp := struct {
		AccessToken string `json:"access_token"`
	}{
		AccessToken: signedToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("AuthService.Handlers.Login: encode response", "error", err)
		http.Error(w, "failed to write response", http.StatusInternalServerError)
		return
	}
}
