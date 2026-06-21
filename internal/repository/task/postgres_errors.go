package task

import (
	"database/sql"
	"errors"
	"task-tracker-1/internal/domain"

	"github.com/jackc/pgx/v5/pgconn"
)

const pgForeignKeyViolation = "23503"

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == pgForeignKeyViolation
}

func mapTaskRepoError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return domain.ErrTaskNotFound
	case isForeignKeyViolation(err):
		return domain.ErrUserNotFound
	default:
		return err
	}
}
