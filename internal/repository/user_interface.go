package repository

import (
	"context"
	"task-tracker-1/internal/domain"
)

type UserRepo interface {
	Save(ctx context.Context, user domain.User) (string, error)
	GetByUsername(ctx context.Context, username string) (domain.User, error)
}
