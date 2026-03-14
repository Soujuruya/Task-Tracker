package service

import (
	"errors"
	"task-tracker-1/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Логика работы с токеном вынесена в отдельный сервис
type TokenService struct {
	jwtSecretKey   []byte
	accessTokenTTL time.Duration
}

type AccessClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewTokenService(jwtSecretKey []byte, accessTokenTTL time.Duration) *TokenService {
	return &TokenService{
		jwtSecretKey:   jwtSecretKey,
		accessTokenTTL: accessTokenTTL,
	}
}

// Добавлена отдельная функция для генерации access-токена
func (t *TokenService) GenerateAccessToken(userID string, username string) (string, error) {
	claims := AccessClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(t.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := accessToken.SignedString(t.jwtSecretKey)
	if err != nil {
		return "", domain.ErrGenerateToken
	}
	return signedToken, nil
}

// Добавлена отдельная функция валидации токена
func (t *TokenService) ValidateAccessToken(token string) (*AccessClaims, error) {
	claims := &AccessClaims{}

	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrTokenInvalid
		}
		return t.jwtSecretKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.ErrTokenExpired
		}
		return nil, domain.ErrTokenInvalid
	}

	return claims, nil
}
