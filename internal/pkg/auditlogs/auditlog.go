package auditlogs

import (
	"time"
)

const (
	ActionUpdate = "update"
	ActionDelete = "delete"
)

const (
	ObjectTypeUser = "user"
	ObjectTypeTask = "task"
)

type AuditLogEntry struct {
	Actor      string    `json:"actor"`       // id редактора
	Action     string    `json:"action"`      // тип операции
	ChangedAt  time.Time `json:"changed_at"`  // время изменения
	ObjectType string    `json:"object_type"` // тип объекта
	ObjectID   string    `json:"object_id"`   // изменяемый объект
	FieldName  string    `json:"field_name"`  // изменяемое поле
	OldValue   string    `json:"old_value"`   // старое значение
	NewValue   string    `json:"new_value"`   // новое значение
}

func NewAuditLogEntry(actor, action, objectType, objectID, fieldName, oldValue, newValue string) *AuditLogEntry {
	entry := AuditLogEntry{
		Actor:      actor,
		Action:     action,
		ChangedAt:  time.Now().UTC(),
		ObjectType: objectType,
		ObjectID:   objectID,
		FieldName:  fieldName,
		OldValue:   oldValue,
		NewValue:   newValue,
	}

	return &entry
}
