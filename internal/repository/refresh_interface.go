package repository

import (
	"context"
	"task-tracker-1/internal/domain"
	"time"
)

type RefreshRepo interface {
	Save(ctx context.Context, token *domain.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	Rotate(ctx context.Context, oldTokenHash string, newToken *domain.RefreshToken, now time.Time) (*domain.RefreshToken, error)
	RevokeByUserID(ctx context.Context, userID string, now time.Time) error
	CleanInvalidTokens(ctx context.Context, now time.Time) error
}
