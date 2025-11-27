package ddd

import "github.com/google/uuid"

// TKey 接口定义了实体的唯一标识符的类型
type TKey interface {
	uint | uint64 | string | uuid.UUID
}

// IEntity 实体具有的能力
type IEntity[T TKey] interface {
	Equal(other IEntity[T]) bool
}

// Entity 具有唯一标识符
type Entity[T TKey] struct {
	ID T
}

// NewEntity 方法用于创建实体
func NewEntity[T TKey](id T) Entity[T] {
	return Entity[T]{
		ID: id,
	}
}

// Equal 方法用于比较两个实体是否相等
func (e *Entity[T]) Equal(other IEntity[T]) bool {
	// 将接口类型转换为具体类型
	if otherEntity, ok := other.(*Entity[T]); ok {
		// 使用 ID 比较实体
		return e.ID == otherEntity.ID
	}

	return false // 如果类型不匹配，返回 false
}
