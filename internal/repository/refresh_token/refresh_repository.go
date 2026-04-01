package refresh_token

import (
	"context"
	"sync"
	"task-tracker-1/internal/domain"
	"time"
)

type RefreshRepository struct {
	mu             sync.RWMutex
	db             map[string]*domain.RefreshToken // tokenID -> one active + all rotate, один ID, но сохраняется история rotate
	tokenHashIdx   map[string]string               // обратный индекс, необходим для получения tokenID по хэшу
	activeTokenIdx map[string]string               // индекст активного токена пользователя userID -> active token
}

func NewRefreshRepository() *RefreshRepository {
	return &RefreshRepository{
		db:             make(map[string]*domain.RefreshToken),
		tokenHashIdx:   make(map[string]string),
		activeTokenIdx: make(map[string]string),
	}
}

func (r *RefreshRepository) Save(ctx context.Context, token *domain.RefreshToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if token == nil {
		return domain.ErrInvalidRefreshTokenState
	}

	// Проверка уникальности tokenID
	if _, exists := r.db[token.ID]; exists {
		return domain.ErrRefreshTokenAlreadyExists
	}
	// Проверка уникальности хэша
	if _, exists := r.tokenHashIdx[token.TokenHash]; exists {
		return domain.ErrRefreshTokenAlreadyExists
	}
	// Проверка, что это не второй активный токен для пользователя
	if token.Status == domain.Active {
		if activeTokenID, exists := r.activeTokenIdx[token.UserID]; exists {
			if activeTokenID != "" {
				return domain.ErrActiveRefreshTokenAlreadyExists
			}
		}
	}
	r.db[token.ID] = token
	r.tokenHashIdx[token.TokenHash] = token.ID

	if token.Status == domain.Active {
		r.activeTokenIdx[token.UserID] = token.ID
	}

	return nil
}

// GetByTokenHash получаем tokenID по хэшу
func (r *RefreshRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tokenID, ok := r.tokenHashIdx[tokenHash]
	if !ok {
		return nil, domain.ErrRefreshTokenNotFound
	}

	token, ok := r.db[tokenID]
	if !ok {
		return nil, domain.ErrRefreshTokenNotFound
	}

	tokenCopy := *token
	return &tokenCopy, nil
}

// Rotate ротация токена; для обеспечения атомарности обноваления токена все выполняется в одной lock секции и реализовано отдельным методом
func (r *RefreshRepository) Rotate(ctx context.Context, oldTokenHash string, newToken *domain.RefreshToken, now time.Time) (*domain.RefreshToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Находим текущий токен по хэшу, который нам пришел
	oldTokenID, ok := r.tokenHashIdx[oldTokenHash]
	if !ok {
		return nil, domain.ErrRefreshTokenNotFound
	}
	// Проверяем, что новый токен вообще можно сохранить в хранилище
	oldToken, ok := r.db[oldTokenID]
	if !ok {
		return nil, domain.ErrRefreshTokenNotFound
	}

	if newToken == nil {
		return nil, domain.ErrInvalidRefreshTokenState
	}
	if newToken.UserID != oldToken.UserID {
		return nil, domain.ErrInvalidRefreshTokenState
	}
	if newToken.Status != domain.Active {
		return nil, domain.ErrInvalidRefreshTokenState
	}
	// Проверяем уникальность ID и хэша нового токена
	if _, exists := r.db[newToken.ID]; exists {
		return nil, domain.ErrRefreshTokenAlreadyExists
	}
	if _, exists := r.tokenHashIdx[newToken.TokenHash]; exists {
		return nil, domain.ErrRefreshTokenAlreadyExists
	}

	// Возвращаем ошибки для уже неактивных токенов,
	// оставляем только для поломанного состояния хранилища.
	if oldToken.Status == domain.Rotated {
		return nil, domain.ErrRefreshTokenRotated
	}
	if oldToken.Status == domain.Revoked {
		return nil, domain.ErrRefreshTokenRevoked
	}
	if !now.Before(oldToken.ExpiresAt) {
		return nil, domain.ErrRefreshTokenExpired
	}

	// Проверяем, что ротируем именно текущий активный токен пользователя
	activeTokenID, ok := r.activeTokenIdx[oldToken.UserID]
	if !ok || activeTokenID != oldToken.ID {
		return nil, domain.ErrRefreshTokenInvalid
	}
	// Переводим старый токен в rotated и связываем с новым токеном
	if err := oldToken.Rotate(newToken.ID, now); err != nil {
		return nil, err
	}
	// Сохраняем и старый и новый токен для истории ротации
	r.db[oldToken.ID] = oldToken
	r.db[newToken.ID] = newToken

	r.tokenHashIdx[newToken.TokenHash] = newToken.ID
	r.activeTokenIdx[newToken.UserID] = newToken.ID

	newTokenCopy := *newToken
	return &newTokenCopy, nil
}

// RevokeByUserID метод принудительной инвалидации токена, для сохранения атомарности операции также выполняются внутри одного lock
func (r *RefreshRepository) RevokeByUserID(ctx context.Context, userID string, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Если у пользователя нет активного токена, считаем revoke не ошибкой, а валидной операцией, так как статус уже правильный
	activeTokenID, ok := r.activeTokenIdx[userID]
	if !ok {
		return nil
	}

	activeToken, ok := r.db[activeTokenID]
	if !ok {
		return domain.ErrInvalidRefreshTokenState
	}

	// Переводим текущий активный токен в состояние revoked
	if err := activeToken.Revoke(now); err != nil {
		return err
	}

	// После revoke у пользователя больше нет активной сессии
	delete(r.activeTokenIdx, userID)

	return nil
}

// CleanInvalidTokens Метод очистки невалидных токенов, запускаем в горутине при старте сервиса
func (r *RefreshRepository) CleanInvalidTokens(ctx context.Context, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for tokenID, token := range r.db {
		if token == nil {
			delete(r.db, tokenID)
			continue
		}
		// Удаляем токены, срок действия которых уже истек
		if !now.Before(token.ExpiresAt) {
			delete(r.db, tokenID)
			delete(r.tokenHashIdx, token.TokenHash)

			if activeTokenID, ok := r.activeTokenIdx[token.UserID]; ok && activeTokenID == tokenID {
				delete(r.activeTokenIdx, token.UserID)
			}
		}
	}

	return nil
}
