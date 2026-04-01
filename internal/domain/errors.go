package domain

import "errors"

// DOMAIN AUTH ERRORS

var (
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidUserID       = errors.New("invalid or missing user id")
	ErrUserNameRequired    = errors.New("username is required")
	ErrPasswordRequired    = errors.New("password is required")
	ErrCredentialsRequired = errors.New("username and password are required")
)

// DOMAIN TASK ERRORS
var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrEmptyTaskTitle    = errors.New("empty task title")
	ErrTaskAlreadyDone   = errors.New("task already done")
	ErrInvalidTaskStatus = errors.New("invalid task status")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrMissingPathID     = errors.New("missing path id")
	ErrInvalidQueryParam = errors.New("invalid query param")
)

// DOMAIN ACCESS TOKEN ERRORS

var (
	ErrTokenExpired  = errors.New("token expired")
	ErrTokenInvalid  = errors.New("token invalid")
	ErrMissingToken  = errors.New("missing token")
	ErrGenerateToken = errors.New("failed to generate token")
)

// DOMAIN REFRESH TOKEN ERRORS
var (
	ErrRefreshTokenExpired             = errors.New("refresh token expired")
	ErrRefreshTokenGenerate            = errors.New("failed to generate refresh token")
	ErrRefreshTokenNotFound            = errors.New("refresh token not found")
	ErrRefreshTokenInvalid             = errors.New("invalid refresh token")
	ErrInvalidRefreshTokenState        = errors.New("invalid refresh token state")
	ErrRefreshTokenRevoked             = errors.New("refresh token revoked")
	ErrRefreshTokenRotated             = errors.New("refresh token rotated")
	ErrRefreshTokenAlreadyExists       = errors.New("refresh token already exists")
	ErrActiveRefreshTokenAlreadyExists = errors.New("active refresh token already exists")
)

// IsTokenDomainError Проверка что ошибка доменная для области токенов
func IsTokenDomainError(err error) bool {
	if err == nil {
		return false
	}
	switch {
	//REFRESH TOKEN
	case errors.Is(err, ErrRefreshTokenExpired):
		return true
	case errors.Is(err, ErrRefreshTokenGenerate):
		return true
	case errors.Is(err, ErrRefreshTokenNotFound):
		return true
	case errors.Is(err, ErrRefreshTokenInvalid):
		return true
	case errors.Is(err, ErrRefreshTokenRevoked):
		return true
	case errors.Is(err, ErrRefreshTokenRotated):
		return true
	case errors.Is(err, ErrRefreshTokenAlreadyExists):
		return true
	case errors.Is(err, ErrActiveRefreshTokenAlreadyExists):
		return true
	case errors.Is(err, ErrInvalidRefreshTokenState):
		return true
	// ACCESS TOKEN
	case errors.Is(err, ErrTokenExpired):
		return true
	case errors.Is(err, ErrTokenInvalid):
		return true
	case errors.Is(err, ErrMissingToken):
		return true
	case errors.Is(err, ErrGenerateToken):
		return true
	default:
		return false
	}
}

func IsAuthDomainError(err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, ErrUserAlreadyExists):
		return true
	case errors.Is(err, ErrUserNotFound):
		return true
	case errors.Is(err, ErrInvalidCredentials):
		return true
	case errors.Is(err, ErrInvalidUserID):
		return true
	case errors.Is(err, ErrUserNameRequired):
		return true
	case errors.Is(err, ErrPasswordRequired):
		return true
	case errors.Is(err, ErrCredentialsRequired):
		return true
	default:
		return false
	}
}
