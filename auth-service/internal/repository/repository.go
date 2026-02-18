package repository

import (
	"auth-service/internal/entity"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)

type UserRepository struct {
	mu sync.RWMutex
	db map[string]entity.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		db: map[string]entity.User{},
	}
}

// Save Сохранение пользователя
func (r *UserRepository) Save(user entity.User) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.db[user.ID]; exists {
		return user.ID, fmt.Errorf("user with id %v already exists: %w", user.ID, ErrUserAlreadyExists)
	}
	r.db[user.ID] = user
	return user.ID, nil
}

// GetByUsername Получение юзера по username для логина
func (r *UserRepository) GetByUsername(username string) (entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Перебор циклом использовал, как более простой рабочий вариант
	for _, user := range r.db {
		if user.Username == username {
			return user, nil
		}
	}
	return entity.User{}, fmt.Errorf("user with username %v not found: %w", username, ErrUserNotFound)
}

// Другие методы не реализованы,т.к. пока нет необходимости их использования
