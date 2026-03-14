package user

import (
	"task-tracker-1/internal/domain"
)

type UserRepo interface {
	Save(user domain.User) (string, error)
	GetByUsername(username string) (domain.User, error)
}
