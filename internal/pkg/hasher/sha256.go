package hasher

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
)

type Sha256Hash struct {
}

func NewSha256Hash() *Sha256Hash {
	return &Sha256Hash{}
}

// Функция для хеширования refresh-токенов
func (h *Sha256Hash) Hash(ctx context.Context, token string) (string, error) {
	hashedToken := sha256.Sum256([]byte(token))

	base64String := base64.RawStdEncoding.EncodeToString(hashedToken[:])

	return base64String, nil
}

// Сравнение хэша пришедшего токена и хэша,который у нас хранится
func (h *Sha256Hash) Compare(ctx context.Context, hash string, token string) error {
	decodedHash, err := base64.RawStdEncoding.DecodeString(hash)
	if err != nil {
		return fmt.Errorf("invalid hash encoding: %w", err)
	}
	tokenHash := sha256.Sum256([]byte(token))
	if len(decodedHash) != len(tokenHash) || subtle.ConstantTimeCompare(decodedHash, tokenHash[:]) != 1 {
		return fmt.Errorf("invalid hash comparison")
	}
	return nil
}
