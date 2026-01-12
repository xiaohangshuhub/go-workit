package ddd

import (
	"reflect"
)

// IValueObject 指对象具有的能力
type IValueObject interface {
	Equal(IValueObject) bool
}

// 值对象
type ValueObject struct{}

// Equals 方法用于动态比较两个值对象是否相等
func (vo ValueObject) Equal(other IValueObject) bool {
	otherValue := reflect.ValueOf(other)

	if otherValue.Kind() != reflect.Ptr || otherValue.IsNil() {
		return false
	}

	otherValue = otherValue.Elem()

	return reflect.DeepEqual(vo, otherValue.Interface())
}
