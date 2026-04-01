package service

import (
	"context"
	"errors"
	"fmt"
	"task-tracker-1/internal/domain"
	"task-tracker-1/internal/pkg"
	"task-tracker-1/internal/pkg/hasher"
	"task-tracker-1/internal/repository"
)

type AuthService struct {
	repo repository.UserRepo
	hash hasher.Hasher
}

func NewAuthService(repo repository.UserRepo, hash hasher.Hasher) *AuthService {
	return &AuthService{
		repo: repo,
		hash: hash,
	}
}

// Register Регистрация + создание пароля и ID
func (s *AuthService) Register(ctx context.Context, username, password string) (string, error) {
	_, err := s.repo.GetByUsername(ctx, username)

	if err == nil {
		return "", domain.ErrUserAlreadyExists
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return "", fmt.Errorf("failed to check existing user: %w", err)
	}

	hash, err := s.hash.Hash(ctx, password)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	userID, err := pkg.GenerateID()
	if err != nil {
		return "", fmt.Errorf("could not generate ID: %w", err)
	}
	newUser := domain.User{
		ID:           userID,
		Username:     username,
		PasswordHash: hash,
	}

	id, err := s.repo.Save(ctx, newUser)
	if err != nil {
		if domain.IsAuthDomainError(err) {
			return "", err
		}
		return "", fmt.Errorf("failed to save user: %w", err)
	}

	return id, nil
}

// Login Логин + проверка пароля
func (s *AuthService) Login(ctx context.Context, username, password string) (domain.User, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.User{}, domain.ErrInvalidCredentials
		}
		return domain.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	if err := s.hash.Compare(ctx, user.PasswordHash, password); err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return domain.User{}, domain.ErrInvalidCredentials
		}
		return domain.User{}, fmt.Errorf("failed to compare password: %w", err)
	}

	return user, nil
}
