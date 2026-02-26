package handlers

import (
	"net/http"
	"task-tracker-1/internal/service"
)

type TaskHandler struct {
	TaskService service.TaskService
}

func NewTaskHandler(taskService service.TaskService) *TaskHandler {
	return &TaskHandler{TaskService: taskService}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {

}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {

}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {

}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {

}
