package handlers

import (
	"log/slog"
	"net/http"
	"task-tracker-1/internal/service"
	"task-tracker-1/internal/transport/dto"
)

type AuthHandler struct {
	AuthService  *service.AuthService
	TokenService *service.TokenService
}

func NewAuthHandler(authService *service.AuthService, tokenService *service.TokenService) *AuthHandler {
	return &AuthHandler{
		AuthService:  authService,
		TokenService: tokenService,
	}
}

// ValidateToken Валидация токена (внутренний эндпоинт)
func (h *AuthHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		slog.Warn("AuthService.Handlers.ValidateToken: method not allowed", "method", r.Method)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestID := getRequestID(r)

	token, err := parseBearerToken(r)
	if writeAuthError(w, requestID, err) {
		return
	}
	slog.Debug("token received", "request_id", requestID)

	accessClaims, err := h.TokenService.ValidateAccessToken(r.Context(), token)
	if writeAuthError(w, requestID, err) {
		return
	}

	resp := dto.ValidateTokenResponse{
		Valid:  true,
		UserID: accessClaims.UserID,
	}

	slog.Debug("token valid", "request_id", requestID, "user_id", accessClaims.UserID)
	if !encodeJSON(w, http.StatusOK, resp) {
		return
	}
}

// Register Регистрация пользователя
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)

	var req dto.RegisterRequest

	defer r.Body.Close()
	if !decodeJSON(w, r, &req) {
		return
	}
	slog.Debug("register attempt", "request_id", requestID, "username", req.Username)

	err := validateRegisterRequest(req)
	if writeAuthError(w, requestID, err) {
		return
	}

	userID, err := h.AuthService.Register(r.Context(), req.Username, req.Password)
	if writeAuthError(w, requestID, err) {
		return
	}
	slog.Info("registration successful", "request_id", requestID, "user_id", userID)

	resp := struct {
		UserID string `json:"user_id"`
	}{
		UserID: userID,
	}

	if !encodeJSON(w, http.StatusCreated, resp) {
		return
	}
}

// Login Аутентификация и получение токена
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)

	var req dto.LoginRequest

	defer r.Body.Close()
	if !decodeJSON(w, r, &req) {
		return
	}
	slog.Debug("login attempt", "request_id", requestID, "username", req.Username)

	err := validateLoginRequest(req)
	if writeAuthError(w, requestID, err) {
		return
	}

	userID, err := h.AuthService.Login(r.Context(), req.Username, req.Password)
	if writeAuthError(w, requestID, err) {
		return
	}

	signedToken, err := h.TokenService.GenerateAccessToken(r.Context(), userID, req.Username)
	if writeAuthError(w, requestID, err) {
		return
	}
	slog.Info("login successful", "request_id", requestID, "username", req.Username, "user_id", userID)

	resp := struct {
		AccessToken string `json:"access_token"`
	}{
		AccessToken: signedToken,
	}

	if !encodeJSON(w, http.StatusOK, resp) {
		return
	}
}
