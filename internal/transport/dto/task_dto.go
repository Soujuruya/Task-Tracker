package dto

type TaskResponse struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	ProgressStatus string `json:"progress_status"`
	CreatedAt      string `json:"created_at"`
}

type CreateTaskRequest struct {
	Title          string `json:"title"`
	Description    string `json:"description"`
	ProgressStatus string `json:"progress_status"`
}

type CreateTaskResponse = TaskResponse

type GetListTasksResponse struct {
	Tasks []TaskResponse `json:"tasks,omitempty"`
}

type UpdateTaskModel struct {
	ID             string  `json:"id"`
	Title          *string `json:"title"`
	Description    *string `json:"description"`
	ProgressStatus *string `json:"progress_status"`
}

type UpdateTaskResponse = TaskResponse
