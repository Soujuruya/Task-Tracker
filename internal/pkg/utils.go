package pkg

import (
	"github.com/google/uuid"
)

func GenerateID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
