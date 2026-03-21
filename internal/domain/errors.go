package domain

import "errors"

var (
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidUserID       = errors.New("invalid or missing user id")
	ErrUserNameRequired    = errors.New("username is required")
	ErrPasswordRequired    = errors.New("password is required")
	ErrCredentialsRequired = errors.New("username and password are required")
)

var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrEmptyTaskTitle    = errors.New("empty task title")
	ErrTaskAlreadyDone   = errors.New("task already done")
	ErrInvalidTaskStatus = errors.New("invalid task status")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrMissingPathID     = errors.New("missing path id")
	ErrInvalidQueryParam = errors.New("invalid query param")
)

var (
	ErrTokenExpired  = errors.New("token expired")
	ErrTokenInvalid  = errors.New("token invalid")
	ErrMissingToken  = errors.New("missing token")
	ErrGenerateToken = errors.New("failed to generate token")
)
