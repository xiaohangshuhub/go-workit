package ddd

import (
	"time"

	"github.com/google/uuid"
)

type IDomainEvent interface {
	Event()
}

type DomainEvent struct {
	EventID uuid.UUID
	Created time.Time
}

// 这是一个空实现，用于标识 DomainEvent
func (d DomainEvent) Event() {

}

// NewDomainEvent 创建一个新的 DomainEvent
func NewDomainEvent() DomainEvent {
	return DomainEvent{
		EventID: uuid.New(),
		Created: time.Now(),
	}
}
