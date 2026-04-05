package middleware

import (
	"context"
	"task-tracker-1/internal/service"
)

type AccessTokenValidator interface {
	ValidateAccessToken(ctx context.Context, token string) (*service.AccessClaims, error)
}
