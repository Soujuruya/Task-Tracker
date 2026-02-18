package service

import (
	"auth-service/internal/entity"
	"auth-service/internal/repository"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Дублирование ошибок, чтобы не тянуть зависимости
var ErrUserAlreadyExists = errors.New("user already exists")

type AuthService struct {
	repo UserRepository
}

// ID генерим здесь
func generateID() (string, error) {
	key := make([]byte, 16)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("failed to generate ID: %w", err)
	}
	return hex.EncodeToString(key), nil
}

func NewAuthService(repo UserRepository) *AuthService {
	return &AuthService{
		repo: repo,
	}
}

// Register Регистрация + создание пароля и ID
func (s *AuthService) Register(username, password string) (string, error) {
	_, err := s.repo.GetByUsername(username)

	if err == nil {
		return "", ErrUserAlreadyExists
	}
	if !errors.Is(err, repository.ErrUserNotFound) {
		return "", fmt.Errorf("failed to check existing user: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("could not hash password: %w", err)
	}

	userID, err := generateID()
	if err != nil {
		return "", fmt.Errorf("could not generate ID: %w", err)
	}
	newUser := entity.User{
		ID:           userID,
		Username:     username,
		PasswordHash: string(hash),
	}

	id, err := s.repo.Save(newUser)
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return "", ErrUserAlreadyExists
		}
		return "", err
	}

	return id, nil
}

// Login Логин + проверка пароля
func (s *AuthService) Login(username, password string) (string, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid password: %w", err)
	}

	return user.ID, nil
}
