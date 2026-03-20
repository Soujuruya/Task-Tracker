package service

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"task-tracker-1/internal/domain"
	"task-tracker-1/internal/pkg"
	"task-tracker-1/internal/pkg/auditlogs"
	"task-tracker-1/internal/pkg/ctxkeys"
	"task-tracker-1/internal/repository"
	"time"
)

type TaskService struct {
	repo     repository.TaskRepo
	auditlog repository.AuditLogRepo
}

func NewTaskService(taskRepo repository.TaskRepo, auditlogRepo repository.AuditLogRepo) *TaskService {
	return &TaskService{
		repo:     taskRepo,
		auditlog: auditlogRepo,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, userID string, task *domain.Task) (string, error) {
	if task.Title == "" {
		return "", domain.ErrEmptyTaskTitle
	}

	if task.ProgressStatus == "" {
		task.ProgressStatus = domain.StatusToDo
	}

	if !domain.IsValidStatus(task.ProgressStatus) {
		return "", domain.ErrInvalidTaskStatus
	}

	taskID, err := pkg.GenerateID()
	if err != nil {
		return "", err
	}

	task.ID = taskID
	task.CreatedAt = time.Now().UTC()

	_, err = s.repo.CreateTask(ctx, userID, task)
	if err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}
	s.auditlog.SaveOwner(ctx, taskID, userID)

	return task.ID, nil
}

func (s *TaskService) GetListTasks(ctx context.Context, userID string) ([]*domain.Task, error) {
	tasks, err := s.repo.GetListTasks(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	//Сортируем таски по времени создании для того чтобы пользователь видел сначала более новые таски
	//Если даты создания идентичны, то сортируем по второму критерию - ID таски.
	slices.SortFunc(tasks, func(i, j *domain.Task) int {
		if i.CreatedAt.Equal(j.CreatedAt) {
			return strings.Compare(i.ID, j.ID)
		}
		return j.CreatedAt.Compare(i.CreatedAt)
	})

	return tasks, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, userID string, task *domain.UpdateTaskInput) (*domain.Task, error) {
	requestID, ok := ctx.Value(ctxkeys.RequestIDKey).(string)
	if !ok {
		slog.Warn("request id not found in context")
	}
	existingTask, err := s.repo.GetTaskByID(ctx, userID, task.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	beforeTask := *existingTask

	if task.Title != nil {
		if *task.Title == "" {
			return nil, domain.ErrEmptyTaskTitle
		}
		existingTask.Title = *task.Title
	}

	if task.Description != nil {
		existingTask.Description = *task.Description
	}

	// Переход статуса
	if task.ProgressStatus != nil {
		if !domain.TryTransition(existingTask.ProgressStatus, *task.ProgressStatus) {
			return nil, domain.ErrInvalidTransition
		}
		slog.Info("task status transition", "request_id", requestID, "user_id", userID, "task_id", task.ID, "from", existingTask.ProgressStatus, "to", *task.ProgressStatus)
		existingTask.ProgressStatus = *task.ProgressStatus
	}

	updatedTask, err := s.repo.UpdateTask(ctx, userID, existingTask)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	auditLog := diffTask(userID, &beforeTask, updatedTask)
	err = s.auditlog.InsertManyAuditLogs(ctx, auditLog)
	// Не выбрасываем ошибку,а просто логируем, так как запрос уже успешный
	if err != nil {
		slog.Error("failed to insert audit logs", "request_id", requestID, "user_id", userID, "task_id", task.ID, "error", err)
	}

	return updatedTask, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, userID, taskID string) error {
	requestID, ok := ctx.Value(ctxkeys.RequestIDKey).(string)
	if !ok {
		slog.Warn("request id not found in context")
	}
	existingTask, err := s.repo.GetTaskByID(ctx, userID, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	if existingTask.ProgressStatus == domain.StatusDone {
		return domain.ErrTaskAlreadyDone
	}

	if err = s.repo.DeleteTask(ctx, userID, taskID); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	entry := auditlogs.NewAuditLogEntry(
		userID, auditlogs.ActionDelete, auditlogs.ObjectTypeTask,
		taskID, "", string(existingTask.ProgressStatus), "",
	)
	// Сохраняем лог удаления задачи
	err = s.auditlog.InsertAuditLog(ctx, entry)
	// Не выбрасываем ошибку,а просто логируем, так как запрос уже успешный
	if err != nil {
		slog.Error("failed to insert audit log", "request_id", requestID, "user_id", userID, "task_id", taskID, "error", err)
	}

	return nil
}

func (s *TaskService) GetTaskHistory(ctx context.Context, userID, taskID string) ([]*auditlogs.AuditLogEntry, error) {
	// Тут проверяем наличие у пользователя данной задачи,чтобы у других не было доступа к ним
	ownerID, err := s.auditlog.GetOwner(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get owner: %w", err)
	}
	if ownerID != userID {
		return nil, domain.ErrTaskNotFound
	}

	entries, err := s.auditlog.GetAllAuditLogs(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return entries, nil
}

// Функция для отслеживания изменения полей таски
func diffTask(actor string, before, after *domain.Task) []*auditlogs.AuditLogEntry {
	var entries []*auditlogs.AuditLogEntry

	add := func(field, oldVal, newVal string) {
		if oldVal != newVal {
			entries = append(entries, auditlogs.NewAuditLogEntry(
				actor, auditlogs.ActionUpdate, auditlogs.ObjectTypeTask,
				before.ID, field, oldVal, newVal,
			))
		}
	}

	add("title", before.Title, after.Title)
	add("description", before.Description, after.Description)
	add("status", string(before.ProgressStatus), string(after.ProgressStatus))

	return entries
}
