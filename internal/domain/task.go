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

type TaskFilter struct {
	Status      *ProgressStatus // фильтрация по статусу
	CreatedFrom *time.Time      // по времени от
	CreatedTo   *time.Time      // по какое время
	Page        int             // номер страницы
	PageSize    int             // кол-во эл-тов на страницу
}

type TaskListResult struct {
	Tasks      []*Task
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

// Функция фильтрации задач по статусу и по времени
func FilterTasks(tasks []*Task, filter TaskFilter) []*Task {
	filteredSlice := make([]*Task, 0)

	for _, task := range tasks {
		if filter.Status != nil && task.ProgressStatus != *filter.Status {
			continue
		}
		if filter.CreatedFrom != nil && task.CreatedAt.Before(*filter.CreatedFrom) {
			continue
		}
		if filter.CreatedTo != nil && task.CreatedAt.After(*filter.CreatedTo) {
			continue
		}
		filteredSlice = append(filteredSlice, task)
	}

	return filteredSlice
}

func TasksPagination(filter TaskFilter, filteredTasks []*Task) (total, start, end int) {
	total = len(filteredTasks)
	start = (filter.Page - 1) * filter.PageSize
	end = start + filter.PageSize

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	return total, start, end
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
