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

	user, err := h.AuthService.Login(r.Context(), req.Username, req.Password)
	if writeAuthError(w, requestID, err) {
		return
	}

	// Теперь три операции объедены в один метод, чтобы логика не была разбросана по хендлеру
	signedToken, refreshToken, err := h.TokenService.IssueLoginTokens(r.Context(), user.ID, user.Username)
	if writeAuthError(w, requestID, err) {
		return
	}
	slog.Info("login successful", "request_id", requestID, "username", user.Username, "user_id", user.ID)

	resp := dto.LoginResponse{
		AccessToken:  signedToken,
		RefreshToken: refreshToken,
	}

	if !encodeJSON(w, http.StatusOK, resp) {
		return
	}
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)
	userID, err := getUserID(r)
	if writeAuthError(w, requestID, err) {
		return
	}
	slog.Debug("logout attempt", "request_id", requestID, "user_id", userID)

	err = h.TokenService.Logout(r.Context(), userID)
	if writeAuthError(w, requestID, err) {
		return
	}
	slog.Info("logout successful", "request_id", requestID, "user_id", userID)

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)

	var req dto.RefreshRequest
	defer r.Body.Close()

	if !decodeJSON(w, r, &req) {
		return
	}
	slog.Debug("refresh attempt", "request_id", requestID)

	err := validateRefreshRequest(req)
	if writeAuthError(w, requestID, err) {
		return
	}

	newAccessToken, newRefreshToken, err := h.TokenService.RefreshToken(r.Context(), req.RefreshToken)
	if writeAuthError(w, requestID, err) {
		return
	}
	slog.Info("token refreshed", "request_id", requestID)

	resp := dto.LoginResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}
	if !encodeJSON(w, http.StatusOK, resp) {
		return
	}
}
