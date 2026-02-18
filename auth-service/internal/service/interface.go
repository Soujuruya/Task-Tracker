package service

import (
	"auth-service/internal/entity"
)

type UserRepository interface {
	Save(user entity.User) (string, error)
	GetByUsername(username string) (entity.User, error)
}
