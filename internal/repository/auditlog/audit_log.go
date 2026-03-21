package auditlog

import (
	"context"
	"sync"
	"task-tracker-1/internal/domain"
	"task-tracker-1/internal/pkg/auditlogs"
)

type AuditLogRepository struct {
	mu       sync.RWMutex
	db       map[string][]*auditlogs.AuditLogEntry
	ownerIdx map[string]string // добавил индекс для получения userID,которому принадлежит задача
}

func NewAuditLogRepository() *AuditLogRepository {
	return &AuditLogRepository{
		db:       make(map[string][]*auditlogs.AuditLogEntry),
		ownerIdx: make(map[string]string),
	}
}

func (r *AuditLogRepository) InsertAuditLog(ctx context.Context, entry *auditlogs.AuditLogEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.db[entry.ObjectID] = append(r.db[entry.ObjectID], entry)
	return nil
}

func (r *AuditLogRepository) GetAllAuditLogs(ctx context.Context, objectID string) ([]*auditlogs.AuditLogEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	auditLogs := r.db[objectID]

	return auditLogs, nil
}

func (r *AuditLogRepository) InsertManyAuditLogs(ctx context.Context, entries []*auditlogs.AuditLogEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, entry := range entries {
		r.db[entry.ObjectID] = append(r.db[entry.ObjectID], entry)
	}
	return nil
}

// SaveOwner GetOwner Добавил методы сохранения userID задачи, чтобы после ее удаления мы не ловили ошибку, что задачи нет
func (r *AuditLogRepository) SaveOwner(ctx context.Context, taskID string, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.ownerIdx[taskID] = userID
	return nil
}

func (r *AuditLogRepository) GetOwner(ctx context.Context, taskID string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ownerIdx, ok := r.ownerIdx[taskID]
	if !ok {
		return "", domain.ErrTaskNotFound
	}
	return ownerIdx, nil
}
