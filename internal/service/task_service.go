package service

import (
	"errors"
	"fmt"
	"slices"
	"task-tracker-1/internal/domain"
	"task-tracker-1/internal/pkg"
	"task-tracker-1/internal/repository/task"
	"time"
)

type TaskService struct {
	repo task.TaskRepo
}

func NewTaskService(repo task.TaskRepo) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) CreateTask(userID string, task *domain.Task) (string, error) {
	if task.Title == "" {
		return "", domain.ErrEmptyTaskTitle
	}

	if task.ProgressStatus == "" {
		task.ProgressStatus = domain.StatusToDo
	} else {
		switch task.ProgressStatus {
		case domain.StatusToDo, domain.StatusInProgress, domain.StatusDone:
		default:
			return "", domain.ErrInvalidTaskStatus
		}
	}

	taskID, err := pkg.GenerateID()
	if err != nil {
		return "", err
	}

	task.ID = taskID
	task.CreatedAt = time.Now()

	_, err = s.repo.CreateTask(userID, task)
	if err != nil {
		return "", err
	}

	return task.ID, nil
}

func (s *TaskService) GetListTasks(userID string) ([]*domain.Task, error) {
	tasks, err := s.repo.GetListTasks(userID)
	if err != nil {
		return nil, err
	}

	//Сортируем таски по времени создании для того чтобы пользователь видел сначала более новые таски,думаю так логичнее
	slices.SortFunc(tasks, func(i, j *domain.Task) int {
		return j.CreatedAt.Compare(i.CreatedAt)
	})

	return tasks, nil
}

func (s *TaskService) UpdateTask(userID string, task *domain.UpdateTaskInput) (*domain.Task, error) {
	existingTask, err := s.repo.GetTaskByID(userID, task.ID)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if task.Title != nil {
		if *task.Title == "" {
			return nil, domain.ErrEmptyTaskTitle
		}
		existingTask.Title = *task.Title
	}
	if task.Description != nil {
		existingTask.Description = *task.Description
	}

	// Логика перехода статусов
	if task.ProgressStatus != nil {
		switch *task.ProgressStatus {
		case domain.StatusDone:
			existingTask.ProgressStatus = domain.StatusDone
		case domain.StatusToDo, domain.StatusInProgress:
			if existingTask.ProgressStatus == domain.StatusDone {
				return nil, fmt.Errorf("impossible to change status from Done: %w", domain.ErrTaskAlreadyDone)
			}
			existingTask.ProgressStatus = *task.ProgressStatus
		default:
			return nil, domain.ErrInvalidTaskStatus
		}
	}

	updatedTask, err := s.repo.UpdateTask(userID, existingTask)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}
	return updatedTask, nil
}

func (s *TaskService) DeleteTask(userID, taskID string) error {
	existingTask, err := s.repo.GetTaskByID(userID, taskID)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			return err
		}
		return fmt.Errorf("failed to get task: %w", err)
	}
	if existingTask.ProgressStatus == domain.StatusDone {
		return fmt.Errorf("impossible to delete task: %w", domain.ErrTaskAlreadyDone)
	}

	if err = s.repo.DeleteTask(userID, taskID); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}
