package repository

import (
	"context"
	"task-tracker-1/internal/domain"
)

type TaskRepo interface {
	CreateTask(ctx context.Context, userID string, task *domain.Task) (string, error)
	GetTaskByID(ctx context.Context, userID, taskID string) (*domain.Task, error)
	GetListTasks(ctx context.Context, userID string) ([]*domain.Task, error)
	UpdateTask(ctx context.Context, userID string, task *domain.Task) (*domain.Task, error)
	DeleteTask(ctx context.Context, userID, taskID string) error
}
