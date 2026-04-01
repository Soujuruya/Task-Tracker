package service

import (
	"context"
	"errors"
	"fmt"
	"task-tracker-1/internal/pkg"
	"task-tracker-1/internal/repository"

	"task-tracker-1/internal/domain"
	"task-tracker-1/internal/pkg/hasher"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Логика работы с токеном вынесена в отдельный сервис
type TokenService struct {
	jwtSecretKey    []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	refreshRepo     repository.RefreshRepo
	hasher          *hasher.Sha256Hash
}

type AccessClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewTokenService(jwtSecretKey []byte, accessTokenTTL time.Duration, refreshTokenTTL time.Duration, repo repository.RefreshRepo, hasher *hasher.Sha256Hash) *TokenService {
	return &TokenService{
		jwtSecretKey:    jwtSecretKey,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		refreshRepo:     repo,
		hasher:          hasher,
	}
}

// Добавлена отдельная функция для генерации access-токена
func (t *TokenService) GenerateAccessToken(ctx context.Context, userID string, username string) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	now := time.Now().UTC()

	claims := AccessClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(t.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := accessToken.SignedString(t.jwtSecretKey)
	if err != nil {
		return "", domain.ErrGenerateToken
	}
	return signedToken, nil
}

// Добавлена отдельная функция валидации access-токена
func (t *TokenService) ValidateAccessToken(ctx context.Context, token string) (*AccessClaims, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

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

func (t *TokenService) GenerateRefreshToken(ctx context.Context, userID string, username string) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	now := time.Now().UTC()

	token, err := pkg.GenerateID()
	if err != nil {
		return "", domain.ErrRefreshTokenGenerate
	}
	tokenID, err := pkg.GenerateID()
	if err != nil {
		return "", domain.ErrRefreshTokenGenerate
	}

	tokenHash, err := t.hasher.Hash(ctx, token)
	if err != nil {
		return "", domain.ErrRefreshTokenGenerate
	}

	refreshToken := domain.RefreshToken{
		ID:        tokenID,
		UserID:    userID,
		Username:  username,
		TokenHash: tokenHash,
		Status:    domain.Active,
		ExpiresAt: now.Add(t.refreshTokenTTL),
		CreatedAt: now,
	}

	if err := refreshToken.Validate(); err != nil {
		return "", err
	}

	if err := t.refreshRepo.Save(ctx, &refreshToken); err != nil {
		if domain.IsTokenDomainError(err) {
			return "", err
		}
		return "", fmt.Errorf("failed to save refresh token: %w", err)
	}

	return token, nil
}

func (t *TokenService) Logout(ctx context.Context, userID string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	now := time.Now().UTC()

	if err := t.refreshRepo.RevokeByUserID(ctx, userID, now); err != nil {
		if domain.IsTokenDomainError(err) {
			return err
		}
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

func (t *TokenService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	if ctx.Err() != nil {
		return "", "", ctx.Err()
	}

	now := time.Now().UTC()

	hashedRefreshToken, err := t.hasher.Hash(ctx, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash refresh token: %w", err)
	}

	oldRefreshToken, err := t.refreshRepo.GetByTokenHash(ctx, hashedRefreshToken)
	if err != nil {
		if domain.IsTokenDomainError(err) {
			return "", "", err
		}
		return "", "", fmt.Errorf("failed to get existing refresh token: %w", err)
	}

	newRefreshToken, err := pkg.GenerateID()
	if err != nil {
		return "", "", domain.ErrRefreshTokenGenerate
	}

	newRefreshTokenID, err := pkg.GenerateID()
	if err != nil {
		return "", "", domain.ErrRefreshTokenGenerate
	}

	hashedNewRefreshToken, err := t.hasher.Hash(ctx, newRefreshToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash new refresh token: %w", err)
	}

	newRefreshTokenEntity := domain.RefreshToken{
		ID:        newRefreshTokenID,
		UserID:    oldRefreshToken.UserID,
		Username:  oldRefreshToken.Username,
		TokenHash: hashedNewRefreshToken,
		Status:    domain.Active,
		CreatedAt: now,
		ExpiresAt: now.Add(t.refreshTokenTTL),
	}

	if err := newRefreshTokenEntity.Validate(); err != nil {
		return "", "", err
	}

	if _, err := t.refreshRepo.Rotate(ctx, hashedRefreshToken, &newRefreshTokenEntity, now); err != nil {
		if domain.IsTokenDomainError(err) {
			return "", "", err
		}
		return "", "", fmt.Errorf("failed to rotate refresh token: %w", err)
	}

	accessToken, err := t.GenerateAccessToken(ctx, oldRefreshToken.UserID, oldRefreshToken.Username)
	if err != nil {
		if domain.IsTokenDomainError(err) {
			return "", "", err
		}
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	return accessToken, newRefreshToken, nil
}
