package ddd

import (
	"time"

	"github.com/google/uuid"
)

type IDomainEvent interface {
	Event()
	EventID() uuid.UUID
	OccurredOn() time.Time
}

type DomainEvent struct {
	EventId uuid.UUID
	Created time.Time
}

// 这是一个空实现，用于标识 DomainEvent
func (d DomainEvent) Event() {

}

func (d DomainEvent) EventID() uuid.UUID {
	return d.EventId
}
func (d DomainEvent) OccurredOn() time.Time {
	return d.Created
}

// NewDomainEvent 创建一个新的 DomainEvent
func NewDomainEvent() DomainEvent {
	return DomainEvent{
		EventId: uuid.New(),
		Created: time.Now(),
	}
}
