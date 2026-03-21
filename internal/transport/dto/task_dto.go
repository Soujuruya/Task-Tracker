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
	Tasks      []TaskResponse `json:"tasks"`
	Pagination Pagination     `json:"pagination"`
}

type Pagination struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalPages int `json:"total_pages"`
}

type UpdateTaskModel struct {
	ID             string  `json:"id"`
	Title          *string `json:"title"`
	Description    *string `json:"description"`
	ProgressStatus *string `json:"progress_status"`
}

type UpdateTaskResponse = TaskResponse

type TaskFilterRequest struct {
	Status      string `json:"status"`
	CreatedFrom string `json:"created_from"`
	CreatedTo   string `json:"created_to"`
	Page        int    `json:"page"`
	PageSize    int    `json:"page_size"`
}
