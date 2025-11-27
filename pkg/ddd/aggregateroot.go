package ddd

// AggregateRoot 聚合根具有的能力
type IAggregateRoot interface {
	AddDomainEvent(event DomainEvent)
	ClearDomainEvents() []DomainEvent
	GetDomainEvents() []DomainEvent
}

// ggregateRoot 聚合根
type AggregateRoot[T TKey] struct {
	Entity[T]
	domainEvents []DomainEvent
}

// NewAggregateRoot 创建聚合根
func NewAggregateRoot[T TKey](id T) AggregateRoot[T] {
	return AggregateRoot[T]{
		Entity:       NewEntity(id),
		domainEvents: make([]DomainEvent, 0),
	}
}

// AddDomainEvent 添加事件
func (ar *AggregateRoot[T]) AddDomainEvent(event DomainEvent) {
	ar.domainEvents = append(ar.domainEvents, event)
}

// ClearDomainEvents 清空事件
func (ar *AggregateRoot[T]) ClearDomainEvents() []DomainEvent {
	events := ar.domainEvents
	ar.domainEvents = make([]DomainEvent, 0)
	return events
}

// GetDomainEvents 获取事件
func (ar *AggregateRoot[T]) GetDomainEvents() []DomainEvent {
	return ar.domainEvents
}
