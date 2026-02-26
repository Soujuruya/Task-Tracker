package domain

import "time"

type ProgressStatus string

const (
	StatusInProgress ProgressStatus = "in_progress"
	StatusDone       ProgressStatus = "done"
	StatusToDo       ProgressStatus = "todo"
)

type Task struct {
	ID             string
	Title          string
	Description    string
	ProgressStatus ProgressStatus
	CreatedAt      time.Time
}

type UpdateTaskInput struct {
	ID             string
	Title          *string
	Description    *string
	ProgressStatus *ProgressStatus
}
