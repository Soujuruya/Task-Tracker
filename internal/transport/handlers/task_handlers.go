package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"task-tracker-1/internal/domain"
	"task-tracker-1/internal/service"
	"task-tracker-1/internal/transport/dto"
	"task-tracker-1/internal/transport/middleware"
	"time"
)

type TaskHandler struct {
	TaskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{TaskService: taskService}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		slog.Error("TaskService.Handlers.CreateTask: unauthorized, missing or invalid userID", "userID", userID)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req dto.CreateTaskRequest
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("TaskService.Handlers.CreateTask: invalid request body", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	task := &domain.Task{
		Title:          req.Title,
		Description:    req.Description,
		ProgressStatus: domain.ProgressStatus(req.ProgressStatus),
	}

	taskID, err := h.TaskService.CreateTask(userID, task)
	if err != nil {
		if errors.Is(err, domain.ErrEmptyTaskTitle) {
			slog.Error("TaskService.Handlers.CreateTask", "error", err)
			http.Error(w, "title is required", http.StatusBadRequest)
			return
		}
		if errors.Is(err, domain.ErrInvalidTaskStatus) {
			slog.Error("TaskService.Handlers.CreateTask", "error", err)
			http.Error(w, "invalid task status", http.StatusBadRequest)
			return
		}
		slog.Error("TaskService.Handlers.CreateTask: internal server error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	slog.Info("TaskService.Handlers.CreateTask: new task created", "taskID", taskID)

	var resp dto.CreateTaskResponse

	resp.ID = taskID
	resp.Title = task.Title
	resp.Description = task.Description
	resp.ProgressStatus = string(task.ProgressStatus)
	resp.CreatedAt = task.CreatedAt.Format(time.RFC3339)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("TaskService.Handlers.CreateTask: encode response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *TaskHandler) GetListTasks(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		slog.Error("TaskService.Handlers.GetListTasks: unauthorized, missing or invalid userID", "userID", userID)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	tasks, err := h.TaskService.GetListTasks(userID)
	if err != nil {
		slog.Error("TaskService.Handlers.GetListTasks: internal server error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	slog.Info("TaskService.Handlers.GetListTasks: new task list")

	taskResponses := make([]dto.TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		taskResponses = append(taskResponses, dto.TaskResponse{
			ID:             task.ID,
			Title:          task.Title,
			Description:    task.Description,
			ProgressStatus: string(task.ProgressStatus),
			CreatedAt:      task.CreatedAt.Format(time.RFC3339),
		})
	}

	resp := dto.GetListTasksResponse{Tasks: taskResponses}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("TaskService.Handlers.GetListTasks: encode response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		slog.Error("TaskService.Handlers.UpdateTask: unauthorized, missing or invalid userID", "userID", userID)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	taskID := r.PathValue("id")
	if taskID == "" {
		slog.Error("TaskService.Handlers.UpdateTask: taskID is required", "taskID", taskID)
		http.Error(w, "task id is required", http.StatusBadRequest)
		return
	}

	var req dto.UpdateTaskModel
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("TaskService.Handlers.UpdateTask: invalid request body", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var progressStatus *domain.ProgressStatus

	if req.ProgressStatus != nil {
		pS := domain.ProgressStatus(*req.ProgressStatus)
		progressStatus = &pS
	}

	reqTask := &domain.UpdateTaskInput{
		ID:             taskID,
		Title:          req.Title,
		Description:    req.Description,
		ProgressStatus: progressStatus,
	}

	updatedTask, err := h.TaskService.UpdateTask(userID, reqTask)
	if err != nil {
		if errors.Is(err, domain.ErrEmptyTaskTitle) {
			slog.Error("TaskService.Handlers.UpdateTask", "error", err)
			http.Error(w, "title is required", http.StatusBadRequest)
			return
		}
		if errors.Is(err, domain.ErrTaskNotFound) {
			slog.Error("TaskService.Handlers.UpdateTask", "error", err)
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, domain.ErrInvalidTaskStatus) {
			slog.Error("TaskService.Handlers.UpdateTask", "error", err)
			http.Error(w, "invalid task status", http.StatusBadRequest)
			return
		}
		if errors.Is(err, domain.ErrInvalidTransition) {
			slog.Error("TaskService.Handlers.UpdateTask", "error", err)
			http.Error(w, "invalid task transition", http.StatusConflict)
			return
		}
		slog.Error("TaskService.Handlers.UpdateTask: internal server error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	slog.Info("TaskService.Handlers.UpdateTask: updated task", "taskID", updatedTask.ID)

	resp := dto.UpdateTaskResponse{
		ID:             taskID,
		Title:          updatedTask.Title,
		Description:    updatedTask.Description,
		ProgressStatus: string(updatedTask.ProgressStatus),
		CreatedAt:      updatedTask.CreatedAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("TaskService.Handlers.UpdateTask: encode response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		slog.Error("TaskService.Handlers.Delete: unauthorized, missing or invalid userID", "userID", userID)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	taskID := r.PathValue("id")
	if taskID == "" {
		slog.Error("TaskService.Handlers.DeleteTask: taskID is required", "taskID", taskID)
		http.Error(w, "task id is required", http.StatusBadRequest)
		return
	}

	err := h.TaskService.DeleteTask(userID, taskID)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			slog.Error("TaskService.Handlers.DeleteTask", "error", err)
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, domain.ErrTaskAlreadyDone) {
			slog.Error("TaskService.Handlers.DeleteTask", "error", err)
			http.Error(w, "can not delete task with status done", http.StatusConflict)
			return
		}
		slog.Error("TaskService.Handlers.DeleteTask: internal server error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	slog.Info("TaskService.Handlers.DeleteTask: deleted task", "taskID", taskID)
	w.WriteHeader(http.StatusNoContent)
}
