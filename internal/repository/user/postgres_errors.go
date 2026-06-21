package user

import (
	"database/sql"
	"errors"
	"task-tracker-1/internal/domain"

	"github.com/jackc/pgx/v5/pgconn"
)

const pgUniqueViolation = "23505"

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == pgUniqueViolation
}

func mapUserRepoError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return domain.ErrUserNotFound
	case isUniqueViolation(err):
		return domain.ErrUserAlreadyExists
	default:
		return err
	}
}
