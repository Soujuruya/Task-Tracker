package domain

import "time"

type ProgressStatus string

// Добавлен статус blocked,описывает ситуацию,когда задача не может быть выполнена из-за какой-либо внешней причины
const (
	StatusInProgress ProgressStatus = "in_progress"
	StatusDone       ProgressStatus = "done"
	StatusToDo       ProgressStatus = "todo"
	StatusBlocked    ProgressStatus = "blocked"
)

// Мапа переходов состояний
var allowedTransitions = map[ProgressStatus][]ProgressStatus{
	StatusToDo:       {StatusInProgress, StatusBlocked},
	StatusInProgress: {StatusDone, StatusBlocked, StatusToDo},
	StatusBlocked:    {StatusToDo, StatusInProgress},
	StatusDone:       {},
}

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

// IsValidStatus Проверяем валидность статуса
func IsValidStatus(status ProgressStatus) bool {
	_, ok := allowedTransitions[status]
	return ok
}

// TryTransition Логика переходов статусов (State Machine)
func TryTransition(from, to ProgressStatus) bool {
	if from == to {
		return true
	}
	for _, allowed := range allowedTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}
