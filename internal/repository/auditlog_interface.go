package repository

import (
	"context"
	"task-tracker-1/internal/pkg/auditlogs"
)

type AuditLogRepo interface {
	InsertAuditLog(ctx context.Context, entry *auditlogs.AuditLogEntry) error
	GetAllAuditLogs(ctx context.Context, objectID string) ([]*auditlogs.AuditLogEntry, error)
	InsertManyAuditLogs(ctx context.Context, entries []*auditlogs.AuditLogEntry) error
	SaveOwner(ctx context.Context, taskID string, userID string)
	GetOwner(ctx context.Context, taskID string) (string, error)
}
