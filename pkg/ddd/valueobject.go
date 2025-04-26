package ddd

import (
	"reflect"
)

// 指对象具有的能力
type IValueObject interface {
	Equal(IValueObject) bool
}

// 值对象
type ValueObject struct{}

// Equals 方法用于动态比较两个值对象是否相等
func (vo ValueObject) Equal(other IValueObject) bool {
	otherValue := reflect.ValueOf(other)

	// 确保其他对象是同类型的值对象
	if otherValue.Kind() != reflect.Ptr || otherValue.IsNil() {
		return false
	}

	// 获取指向值对象的值
	otherValue = otherValue.Elem()

	// 进行比较
	return reflect.DeepEqual(vo, otherValue.Interface())
}
