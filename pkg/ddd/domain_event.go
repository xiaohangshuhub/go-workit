package ddd

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent
type DomainEvent struct {
	EventId   uuid.UUID
	Created   time.Time
	EventName string
}
