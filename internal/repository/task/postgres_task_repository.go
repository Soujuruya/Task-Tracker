package task

import (
	"context"
	"database/sql"
	"task-tracker-1/internal/domain"
)

type PostgresTaskRepository struct {
	db *sql.DB
}

func NewPostgresTaskRepository(db *sql.DB) *PostgresTaskRepository {
	return &PostgresTaskRepository{
		db: db,
	}
}

func (r *PostgresTaskRepository) CreateTask(ctx context.Context, userID string, task *domain.Task) (string, error) {
	const query = `
		INSERT INTO tasks (id,user_id,title,description,progress_status,created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		task.ID,
		userID,
		task.Title,
		task.Description,
		task.ProgressStatus,
		task.CreatedAt,
	)
	if err != nil {
		return "", mapTaskRepoError(err)
	}
	return task.ID, nil
}

func (r *PostgresTaskRepository) GetTaskByID(ctx context.Context, userID, taskID string) (*domain.Task, error) {
	const query = `
	SELECT id, title, description, progress_status, created_at
	FROM tasks
	WHERE id = $1 AND user_id = $2
	`
	var task domain.Task

	row := r.db.QueryRowContext(ctx, query, taskID, userID)
	err := row.Scan(&task.ID, &task.Title, &task.Description, &task.ProgressStatus, &task.CreatedAt)
	if err != nil {
		return nil, mapTaskRepoError(err)
	}

	return &task, nil
}

func (r *PostgresTaskRepository) GetListTasks(ctx context.Context, userID string) ([]*domain.Task, error) {
	const query = `
			SELECT id, title, description, progress_status, created_at
			FROM tasks
			WHERE user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, mapTaskRepoError(err)
	}
	defer rows.Close()

	var tasks []*domain.Task

	for rows.Next() {
		var task domain.Task

		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.ProgressStatus, &task.CreatedAt); err != nil {
			return nil, mapTaskRepoError(err)
		}

		tasks = append(tasks, &task)
	}
	if err := rows.Err(); err != nil {
		return nil, mapTaskRepoError(err)
	}
	return tasks, nil
}

func (r *PostgresTaskRepository) UpdateTask(ctx context.Context, userID string, task *domain.Task) (*domain.Task, error) {
	const query = `
  		UPDATE tasks
  		SET title = $3,
  		    description = $4,
  		    progress_status = $5
  		WHERE id = $1 AND user_id = $2
  		RETURNING id, title, description, progress_status, created_at
  	`

	var updatedTask domain.Task

	row := r.db.QueryRowContext(
		ctx,
		query,
		task.ID,
		userID,
		task.Title,
		task.Description,
		task.ProgressStatus,
	)
	err := row.Scan(
		&updatedTask.ID,
		&updatedTask.Title,
		&updatedTask.Description,
		&updatedTask.ProgressStatus,
		&updatedTask.CreatedAt,
	)
	if err != nil {
		return nil, mapTaskRepoError(err)
	}
	return &updatedTask, nil
}

func (r *PostgresTaskRepository) DeleteTask(ctx context.Context, userID string, taskID string) error {
	const query = `
		DELETE FROM tasks
		WHERE id = $1 AND user_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, taskID, userID)
	if err != nil {
		return mapTaskRepoError(err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return mapTaskRepoError(err)
	}
	if affected == 0 {
		return domain.ErrTaskNotFound
	}
	return nil
}
