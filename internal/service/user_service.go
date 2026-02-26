package service

import (
	"errors"
	"fmt"
	"task-tracker-1/internal/domain"
	"task-tracker-1/internal/pkg"
	"task-tracker-1/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo repository.UserRepo
}

func NewAuthService(repo repository.UserRepo) *AuthService {
	return &AuthService{
		repo: repo,
	}
}

// Register Регистрация + создание пароля и ID
func (s *AuthService) Register(username, password string) (string, error) {
	_, err := s.repo.GetByUsername(username)

	if err == nil {
		return "", domain.ErrUserAlreadyExists
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return "", fmt.Errorf("failed to check existing user: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("could not hash password: %w", err)
	}

	userID, err := pkg.GenerateID()
	if err != nil {
		return "", fmt.Errorf("could not generate ID: %w", err)
	}
	newUser := domain.User{
		ID:           userID,
		Username:     username,
		PasswordHash: string(hash),
	}

	id, err := s.repo.Save(newUser)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			return "", err
		}
		return "", fmt.Errorf("failed to save user: %w", err)
	}

	return id, nil
}

// Login Логин + проверка пароля
func (s *AuthService) Login(username, password string) (string, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", domain.ErrInvalidCredentials
		}
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", domain.ErrInvalidCredentials
	}

	return user.ID, nil
}
