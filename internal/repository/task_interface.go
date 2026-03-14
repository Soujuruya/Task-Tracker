package task

import "task-tracker-1/internal/domain"

type TaskRepo interface {
	CreateTask(userID string, task *domain.Task) (string, error)
	GetTaskByID(userID, taskID string) (*domain.Task, error)
	GetListTasks(userID string) ([]*domain.Task, error)
	UpdateTask(userID string, task *domain.Task) (*domain.Task, error)
	DeleteTask(userID, taskID string) error
}
