package domain

import "errors"

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrEmptyTaskTitle    = errors.New("empty task title")
	ErrTaskAlreadyDone   = errors.New("task already done")
	ErrInvalidTaskStatus = errors.New("invalid task status")
	ErrInvalidTransition = errors.New("invalid status transition")
)
