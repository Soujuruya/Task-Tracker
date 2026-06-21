package user

import (
	"context"
	"database/sql"
	"task-tracker-1/internal/domain"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

func (r *PostgresUserRepository) Save(ctx context.Context, user domain.User) (string, error) {
	const query = `
  		INSERT INTO users (id, username, password_hash)
  		VALUES ($1, $2, $3)
  	`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Username, user.PasswordHash)
	if err != nil {
		return "", mapUserRepoError(err)
	}
	return user.ID, nil
}

func (r *PostgresUserRepository) GetByUsername(ctx context.Context, username string) (domain.User, error) {
	const query = `
    		SELECT id, username, password_hash 
    		FROM users 
    		WHERE username = $1
	`

	var user domain.User

	row := r.db.QueryRowContext(ctx, query, username)
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		return domain.User{}, mapUserRepoError(err)
	}

	return user, nil
}
