package user

import (
	"fmt"
	"sync"
	"task-tracker-1/internal/domain"
)

type UserRepository struct {
	mu sync.RWMutex
	db map[string]domain.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		db: map[string]domain.User{},
	}
}

// Save Сохранение пользователя
func (r *UserRepository) Save(user domain.User) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.db[user.ID]; exists {
		return user.ID, fmt.Errorf("user with id %v already exists: %w", user.ID, domain.ErrUserAlreadyExists)
	}
	r.db[user.ID] = user
	return user.ID, nil
}

// GetByUsername Получение юзера по username для логина
func (r *UserRepository) GetByUsername(username string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Перебор циклом использовал, как более простой рабочий вариант
	for _, user := range r.db {
		if user.Username == username {
			return user, nil
		}
	}
	return domain.User{}, fmt.Errorf("user with username %v not found: %w", username, domain.ErrUserNotFound)
}

// Другие методы не реализованы,т.к. пока нет необходимости их использования
