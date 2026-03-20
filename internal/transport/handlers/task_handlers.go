package handlers

import (
	"log/slog"
	"net/http"
	"task-tracker-1/internal/domain"
	"task-tracker-1/internal/service"
	"task-tracker-1/internal/transport/dto"
	"time"
)

type TaskHandler struct {
	TaskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{TaskService: taskService}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)

	userID, err := getUserID(r)
	if writeTaskError(w, requestID, err) {
		return
	}

	var req dto.CreateTaskRequest
	defer r.Body.Close()

	if !decodeJSON(w, r, &req) {
		return
	}

	slog.Debug("create task request", "request_id", requestID, "title", req.Title, "status", req.ProgressStatus)

	task := &domain.Task{
		Title:          req.Title,
		Description:    req.Description,
		ProgressStatus: domain.ProgressStatus(req.ProgressStatus),
	}

	taskID, err := h.TaskService.CreateTask(r.Context(), userID, task)
	if writeTaskError(w, requestID, err) {
		return
	}
	slog.Info("task created", "request_id", requestID, "user_id", userID, "task_id", taskID)

	var resp dto.CreateTaskResponse

	resp.ID = taskID
	resp.Title = task.Title
	resp.Description = task.Description
	resp.ProgressStatus = string(task.ProgressStatus)
	resp.CreatedAt = task.CreatedAt.Format(time.RFC3339)

	if !encodeJSON(w, http.StatusCreated, resp) {
		return
	}
}

func (h *TaskHandler) GetListTasks(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)
	userID, err := getUserID(r)
	if writeTaskError(w, requestID, err) {
		return
	}

	tasks, err := h.TaskService.GetListTasks(r.Context(), userID)
	if writeTaskError(w, requestID, err) {
		return
	}
	slog.Info("new task list", "request_id", requestID, "user_id", userID, "tasks_count", len(tasks))

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

	if !encodeJSON(w, http.StatusOK, resp) {
		return
	}
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)
	userID, err := getUserID(r)
	if writeTaskError(w, requestID, err) {
		return
	}

	taskID, err := getPathID(r)
	if writeTaskError(w, requestID, err) {
		return
	}

	var req dto.UpdateTaskModel
	defer r.Body.Close()

	if !decodeJSON(w, r, &req) {
		return
	}

	slog.Debug("update task request", "request_id", requestID, "task_id", taskID, "title", req.Title, "status", req.ProgressStatus)

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

	updatedTask, err := h.TaskService.UpdateTask(r.Context(), userID, reqTask)
	if writeTaskError(w, requestID, err) {
		return
	}
	slog.Info("task updated", "request_id", requestID, "user_id", userID, "task_id", taskID)

	resp := dto.UpdateTaskResponse{
		ID:             taskID,
		Title:          updatedTask.Title,
		Description:    updatedTask.Description,
		ProgressStatus: string(updatedTask.ProgressStatus),
		CreatedAt:      updatedTask.CreatedAt.Format(time.RFC3339),
	}

	if !encodeJSON(w, http.StatusOK, resp) {
		return
	}
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)
	userID, err := getUserID(r)
	if writeTaskError(w, requestID, err) {
		return
	}

	taskID, err := getPathID(r)
	if writeTaskError(w, requestID, err) {
		return
	}

	err = h.TaskService.DeleteTask(r.Context(), userID, taskID)
	if writeTaskError(w, requestID, err) {
		return
	}
	slog.Info("task deleted", "request_id", requestID, "user_id", userID, "task_id", taskID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) GetTaskHistory(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r)
	userID, err := getUserID(r)
	if writeTaskError(w, requestID, err) {
		return
	}

	taskID, err := getPathID(r)
	if writeTaskError(w, requestID, err) {
		return
	}

	auditLogs, err := h.TaskService.GetTaskHistory(r.Context(), userID, taskID)
	if writeTaskError(w, requestID, err) {
		return
	}
	slog.Info("task history", "request_id", requestID, "task_id", taskID, "changes", len(auditLogs))
	if !encodeJSON(w, http.StatusOK, auditLogs) {
		return
	}
}
