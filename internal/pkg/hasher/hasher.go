package hasher

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"task-tracker-1/internal/domain"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

type Hasher interface {
	Hash(password string) (string, error)
	Compare(hash string, password string) error
}

type Argon2Params struct {
	Memory      uint32 //сколько памяти используется алгоритм в килобайтах
	Iterations  uint32 // кол-во проходов по памяти
	Parallelism uint8  // кол-во параллельных потоков
	SaltLength  uint32 // длина случайной соли(случайный набор байт)
	KeyLength   uint32 // длина итогового хэша в байтах
}

type Argon2Hasher struct {
	parameters Argon2Params
}

func NewArgon2Hasher(parameters Argon2Params) *Argon2Hasher {
	return &Argon2Hasher{parameters: parameters}
}

// Hash Хэширование пароля
func (h Argon2Hasher) Hash(password string) (string, error) {
	salt, err := generateRandomBytes(h.parameters.SaltLength)
	if err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.parameters.Iterations,
		h.parameters.Memory,
		h.parameters.Parallelism,
		h.parameters.KeyLength,
	)

	// Кодируем, чтобы хранить как строку
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Собираем все в нужный формат
	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.parameters.Memory,
		h.parameters.Iterations,
		h.parameters.Parallelism,
		b64Salt,
		b64Hash,
	)

	return encoded, nil
}

// Compare Сравнение пароля и хэша
func (h Argon2Hasher) Compare(encodedHash string, password string) error {
	// Парсим соль и хэш
	salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return fmt.Errorf("failed to decode hash: %w", err)
	}

	// Хэшируем пароль с той же солью и параметрами
	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		h.parameters.Iterations,
		h.parameters.Memory,
		h.parameters.Parallelism,
		h.parameters.KeyLength,
	)

	// Сравниваем байты
	// Используем именно ConstantTimeCompare, чтобы время проверки всегда одинаковое было
	if subtle.ConstantTimeCompare(hash, otherHash) != 1 {
		return domain.ErrInvalidCredentials
	}

	return nil
}

// Генерация соли(случайного набора байт)
func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

// Декодируем хэш ($argon2id$v=19$m=65536,t=3,p=2$Woo1mErn1s7AHf96ewQ8Uw$D4TzIwGO4XD2buk96qAP+Ed2baMo/KbTRMqXX00wtsU)
func decodeHash(encodedHash string) (salt, hash []byte, err error) {
	// Разбиваем на пустую строку,argon2id, версию, параметры, соль, хэш
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, ErrInvalidHash
	}

	// Проверяем версию алгоритма
	// Если хэш создан старой версией Argon2, то ошибка
	var version int
	if _, err = fmt.Sscanf(vals[2], "v=%d", &version); err != nil {
		return nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, ErrIncompatibleVersion
	}

	// Декодируем соль
	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode salt: %w", err)
	}
	// Декодируем хэш
	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode hash: %w", err)
	}

	return salt, hash, nil
}
