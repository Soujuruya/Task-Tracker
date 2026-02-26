package pkg

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateID() (string, error) {
	key := make([]byte, 16)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("failed to generate ID: %w", err)
	}
	return hex.EncodeToString(key), nil
}
