package ddd

// 聚合根具有的能力
type IAggregateRoot interface {
	AddDomainEvent(event IDomainEvent)
	ClearDomainEvents() []IDomainEvent
	GetDomainEvents() []IDomainEvent
}

// 聚合根
type AggregateRoot[T TKey] struct {
	Entity[T]
	domainEvents []IDomainEvent
}

func NewAggregateRoot[T TKey](id T) AggregateRoot[T] {
	return AggregateRoot[T]{
		Entity:       NewEntity(id),
		domainEvents: make([]IDomainEvent, 0),
	}
}

// 添加事件
func (ar *AggregateRoot[T]) AddDomainEvent(event IDomainEvent) {
	ar.domainEvents = append(ar.domainEvents, event)
}

// 清空事件
func (ar *AggregateRoot[T]) ClearDomainEvents() []IDomainEvent {
	events := ar.domainEvents
	ar.domainEvents = make([]IDomainEvent, 0)
	return events
}

// 获取事件
func (ar *AggregateRoot[T]) GetDomainEvents() []IDomainEvent {
	return ar.domainEvents
}
