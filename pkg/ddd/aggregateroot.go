package ddd

// AggregateRoot 聚合根具有的能力
type AggregateRoot interface {
	AddDomainEvent(event DomainEvent)
	ClearDomainEvents() []DomainEvent
	GetDomainEvents() []DomainEvent
}

// BaseAggregateRoot 聚合根
type BaseAggregateRoot[T TKey] struct {
	Entity[T]
	domainEvents []DomainEvent
}

// NewBaseAggregateRoot
func NewBaseAggregateRoot[T TKey](id T) BaseAggregateRoot[T] {
	return BaseAggregateRoot[T]{
		Entity:       NewEntity(id),
		domainEvents: make([]DomainEvent, 0),
	}
}

// AddDomainEvent 添加事件
func (ar *BaseAggregateRoot[T]) AddDomainEvent(event DomainEvent) {
	ar.domainEvents = append(ar.domainEvents, event)
}

// ClearDomainEvents 清空事件
func (ar *BaseAggregateRoot[T]) ClearDomainEvents() []DomainEvent {
	events := ar.domainEvents
	ar.domainEvents = make([]DomainEvent, 0)
	return events
}

// GetDomainEvents 获取事件
func (ar *BaseAggregateRoot[T]) GetDomainEvents() []DomainEvent {
	return ar.domainEvents
}
