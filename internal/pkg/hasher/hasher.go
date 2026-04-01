package hasher

import (
	"context"
	"errors"
)

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

type Hasher interface {
	Hash(ctx context.Context, value string) (string, error)
	Compare(ctx context.Context, hash string, password string) error
}
