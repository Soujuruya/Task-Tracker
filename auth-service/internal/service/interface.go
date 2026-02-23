package service

import (
	"auth-service/internal/domain"
)

type UserRepository interface {
	Save(user domain.User) (string, error)
	GetByUsername(username string) (domain.User, error)
}
