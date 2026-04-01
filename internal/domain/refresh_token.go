package domain

import "time"

type TokenStatus string

const (
	Active  TokenStatus = "active"
	Rotated TokenStatus = "rotated"
	Revoked TokenStatus = "revoked"
)

type RefreshToken struct {
	ID                string
	UserID            string
	Username          string
	TokenHash         string
	Status            TokenStatus
	ExpiresAt         time.Time
	CreatedAt         time.Time
	InvalidatedAt     *time.Time
	ReplacedByTokenID *string
}

// Validate валидация консистености структуры токена
func (t *RefreshToken) Validate() error {
	if t == nil {
		return ErrInvalidRefreshTokenState
	}
	if t.ID == "" || t.UserID == "" || t.TokenHash == "" || t.Username == "" {
		return ErrInvalidRefreshTokenState
	}
	if t.CreatedAt.IsZero() || t.ExpiresAt.IsZero() {
		return ErrInvalidRefreshTokenState
	}
	if !t.ExpiresAt.After(t.CreatedAt) {
		return ErrInvalidRefreshTokenState
	}

	switch t.Status {
	case Active:
		if t.InvalidatedAt != nil || t.ReplacedByTokenID != nil {
			return ErrInvalidRefreshTokenState
		}
		return nil

	case Rotated:
		// Проверяем,что и под указателем, чтобы нельзя было подсунуть пустой токен
		if t.InvalidatedAt == nil || t.ReplacedByTokenID == nil || *t.ReplacedByTokenID == "" {
			return ErrInvalidRefreshTokenState
		}
		return nil

	case Revoked:
		if t.InvalidatedAt == nil || t.ReplacedByTokenID != nil {
			return ErrInvalidRefreshTokenState
		}
		return nil

	default:
		return ErrInvalidRefreshTokenState
	}
}

// Rotate метод, который позволяет штатно заменить старый токен на новый
func (t *RefreshToken) Rotate(newTokenID string, now time.Time) error {
	now = now.UTC()

	if err := t.Validate(); err != nil {
		return err
	}
	if newTokenID == "" {
		return ErrInvalidRefreshTokenState
	}

	switch t.Status {
	case Rotated:
		return ErrRefreshTokenRotated
	case Revoked:
		return ErrRefreshTokenRevoked
	case Active:
	default:
		return ErrInvalidRefreshTokenState
	}

	if t.isExpired(now) {
		return ErrRefreshTokenExpired
	}

	t.InvalidatedAt = &now
	t.ReplacedByTokenID = &newTokenID
	t.Status = Rotated

	return t.Validate()
}

// Revoke метод, который принудительно инвалидирует токен
// Проверяем валидность структуры токена и в начале и в конце, чтобы не нарушать констистентность
func (t *RefreshToken) Revoke(now time.Time) error {
	now = now.UTC()

	if err := t.Validate(); err != nil {
		return err
	}

	switch t.Status {
	case Rotated:
		return ErrRefreshTokenRotated
	case Revoked:
		// По сути если статус уже Revoked, то мы пришли к правильному состоянию и нет смысла возращать ошибку
		return nil
	case Active:
	default:
		return ErrInvalidRefreshTokenState
	}

	// Также истекший токен уже не валиден, поэтому не возращаем ошибку
	if t.isExpired(now) {
		return nil
	}

	t.InvalidatedAt = &now
	t.Status = Revoked

	return t.Validate()
}

func (t *RefreshToken) isExpired(now time.Time) bool {
	if now.Before(t.ExpiresAt) {
		return false
	}
	return true
}
