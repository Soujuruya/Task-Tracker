package task

import (
	"context"
	"sync"
	"task-tracker-1/internal/domain"
)

type MemoryTaskRepository struct {
	mu sync.RWMutex
	db map[string]map[string]*domain.Task
}

// Дополнительный метод, который позвляет передавать копию таски, чтобы сервис не изменял их вне lock секции в map
func cloneTask(task *domain.Task) *domain.Task {
	if task == nil {
		return nil
	}
	taskCopy := *task
	return &taskCopy
}

func NewMemoryTaskRepository() *MemoryTaskRepository {
	return &MemoryTaskRepository{
		db: make(map[string]map[string]*domain.Task),
	}
}

func (r *MemoryTaskRepository) CreateTask(ctx context.Context, userID string, task *domain.Task) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.db[userID] == nil {
		r.db[userID] = make(map[string]*domain.Task)
	}
	r.db[userID][task.ID] = cloneTask(task)

	return task.ID, nil
}

func (r *MemoryTaskRepository) GetTaskByID(ctx context.Context, userID, taskID string) (*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.db[userID][taskID]
	if !ok {
		return nil, domain.ErrTaskNotFound
	}

	return cloneTask(task), nil
}

func (r *MemoryTaskRepository) GetListTasks(ctx context.Context, userID string) ([]*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := r.db[userID]

	tasksSlice := make([]*domain.Task, 0, len(tasks))
	for _, task := range tasks {
		tasksSlice = append(tasksSlice, cloneTask(task))
	}
	return tasksSlice, nil
}

func (r *MemoryTaskRepository) UpdateTask(ctx context.Context, userID string, task *domain.Task) (*domain.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.db[userID]; !ok {
		return nil, domain.ErrUserNotFound
	}

	if _, ok := r.db[userID][task.ID]; !ok {
		return nil, domain.ErrTaskNotFound
	}

	r.db[userID][task.ID] = cloneTask(task)

	return cloneTask(r.db[userID][task.ID]), nil
}

func (r *MemoryTaskRepository) DeleteTask(ctx context.Context, userID string, taskID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.db[userID]; !ok {
		return domain.ErrUserNotFound
	}

	if _, ok := r.db[userID][taskID]; !ok {
		return domain.ErrTaskNotFound
	}

	delete(r.db[userID], taskID)

	return nil
}
