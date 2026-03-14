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
	userID, err := getUserID(r)
	if writeTaskError(w, "TaskService.Handlers.CreateTask", err) {
		return
	}

	var req dto.CreateTaskRequest
	defer r.Body.Close()

	if !decodeJSON(w, r, &req) {
		return
	}

	task := &domain.Task{
		Title:          req.Title,
		Description:    req.Description,
		ProgressStatus: domain.ProgressStatus(req.ProgressStatus),
	}

	taskID, err := h.TaskService.CreateTask(userID, task)
	if writeTaskError(w, "TaskService.Handlers.CreateTask", err) {
		return
	}
	slog.Info("TaskService.Handlers.CreateTask: new task created", "taskID", taskID)

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
	userID, err := getUserID(r)
	if writeTaskError(w, "TaskService.Handlers.GetListTasks", err) {
		return
	}

	tasks, err := h.TaskService.GetListTasks(userID)
	if writeTaskError(w, "TaskService.Handlers.GetListTasks", err) {
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

	if !encodeJSON(w, http.StatusOK, resp) {
		return
	}
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if writeTaskError(w, "TaskService.Handlers.UpdateTask", err) {
		return
	}

	taskID, err := getPathID(r)
	if writeTaskError(w, "TaskService.Handlers.UpdateTask", err) {
		return
	}

	var req dto.UpdateTaskModel
	defer r.Body.Close()

	if !decodeJSON(w, r, &req) {
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
	if writeTaskError(w, "TaskService.Handlers.UpdateTask", err) {
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

	if !encodeJSON(w, http.StatusOK, resp) {
		return
	}
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if writeTaskError(w, "TaskService.Handlers.DeleteTask", err) {
		return
	}

	taskID, err := getPathID(r)
	if writeTaskError(w, "TaskService.Handlers.DeleteTask", err) {
		return
	}

	err = h.TaskService.DeleteTask(userID, taskID)
	if writeTaskError(w, "TaskService.Handlers.DeleteTask", err) {
		return
	}
	slog.Info("TaskService.Handlers.DeleteTask: deleted task", "taskID", taskID)
	w.WriteHeader(http.StatusNoContent)
}
