package user

import (
	"sync"
	"task-tracker-1/internal/domain"
)

type UserRepository struct {
	mu          sync.RWMutex
	db          map[string]domain.User
	usernameIdx map[string]string
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		db:          map[string]domain.User{},
		usernameIdx: map[string]string{},
	}
}

// Save Сохранение пользователя
func (r *UserRepository) Save(user domain.User) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.db[user.ID]; exists {
		return "", domain.ErrUserAlreadyExists
	}
	//Вторичный индекс для быстрого поиска
	if _, exists := r.usernameIdx[user.Username]; exists {
		return "", domain.ErrUserAlreadyExists
	}
	r.db[user.ID] = user
	r.usernameIdx[user.Username] = user.ID
	return user.ID, nil
}

// GetByUsername Получение юзера по username для логина
func (r *UserRepository) GetByUsername(username string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Поиск работает теперь по вторичному индексу O(1)
	userID, ok := r.usernameIdx[username]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}

	return r.db[userID], nil
}
