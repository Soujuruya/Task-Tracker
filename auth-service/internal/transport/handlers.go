package transport

import (
	"auth-service/internal/domain"
	"auth-service/internal/service"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
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
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ValidateTokenRequest

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Handlers:ValidateToken: invalid request body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	log.Printf("Handlers:ValidateToken: token received")

	var accessClaims struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		jwt.RegisteredClaims
	}

	_, err := jwt.ParseWithClaims(req.Token, &accessClaims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return h.JWTSecretKey, nil
	})

	if err != nil {
		log.Printf("Handlers:ValidateToken: failed to parse token: %v", err)
		if errors.Is(err, jwt.ErrTokenExpired) {
			http.Error(w, "token expired", http.StatusUnauthorized)
			return
		}

		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	resp := ValidateTokenResponse{
		Valid:  true,
		UserID: accessClaims.UserID,
	}

	log.Printf("Handlers:ValidateToken: token valid for user_id: %s", accessClaims.UserID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
		return
	}
}

// Register Регистрация пользователя
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Handlers:Register: invalid request body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Handlers:Register: registration attempt: username=%s", req.Username)

	if req.Username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}

	userID, err := h.AuthService.Register(req.Username, req.Password)
	if err != nil {
		log.Printf("Handlers:Register: registration failed: %v", err)
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			http.Error(w, "user already exists", http.StatusConflict)
			return
		}
		log.Printf("Handlers:Register: internal error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Handlers:Register: registration successful: user_id=%s", userID)
	resp := struct {
		UserID string `json:"user_id"`
	}{
		UserID: userID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
		return
	}
}

// Login Аутентификация и получение токена
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Handlers:Login: invalid request body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Handlers:Login: login attempt: username=%s", req.Username)

	if req.Username == "" || req.Password == "" {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	userID, err := h.AuthService.Login(req.Username, req.Password)

	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		log.Printf("Handlers:Login: login failed for username=%s: %v", req.Username, err)
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
		log.Printf("Handlers:Login: failed to generate token for username=%s: %v", req.Username, err)
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	log.Printf("Handlers:Login: login successful: username=%s, user_id=%s", req.Username, userID)
	resp := struct {
		AccessToken string `json:"access_token"`
	}{
		AccessToken: signedToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
		return
	}
}
