package task

import (
	"fmt"
	"sync"
	"task-tracker-1/internal/domain"
)

type TaskRepository struct {
	mu sync.RWMutex
	db map[string]map[string]*domain.Task
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		db: make(map[string]map[string]*domain.Task),
	}
}

func (r *TaskRepository) CreateTask(userID string, task *domain.Task) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.db[userID] == nil {
		r.db[userID] = make(map[string]*domain.Task)
	}
	r.db[userID][task.ID] = task

	return task.ID, nil
}

func (r *TaskRepository) GetTaskByID(userID, taskID string) (*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.db[userID][taskID]
	if !ok {
		return nil, fmt.Errorf("task with id %s not found: %w", taskID, domain.ErrTaskNotFound)
	}

	return task, nil
}

func (r *TaskRepository) GetListTasks(userID string) ([]*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := r.db[userID]

	tasksSlice := make([]*domain.Task, 0, len(tasks))
	for _, task := range tasks {
		tasksSlice = append(tasksSlice, task)
	}
	return tasksSlice, nil
}

func (r *TaskRepository) UpdateTask(userID string, task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.db[userID][task.ID]; !ok {
		return fmt.Errorf("task with id %s not found: %w", task.ID, domain.ErrTaskNotFound)
	}

	r.db[userID][task.ID] = task
	return nil
}

func (r *TaskRepository) DeleteTask(userID string, taskID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.db[userID][taskID]; !ok {
		return fmt.Errorf("task with id %s not found: %w", taskID, domain.ErrTaskNotFound)
	}

	delete(r.db[userID], taskID)

	return nil
}
