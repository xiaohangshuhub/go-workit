package collections

// 检查集合是否包含某个元素
func Contains[T comparable](collection []T, item T) bool {
	for _, v := range collection {
		if v == item {
			return true
		}
	}
	return false
}

// 移除集合中的重复元素
func RemoveDuplicates[T comparable](collection []T) []T {
	unique := make(map[T]struct{})
	result := make([]T, 0)
	for _, v := range collection {
		if _, exists := unique[v]; !exists {
			unique[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// 过滤集合
func Filter[T any](collection []T, predicate func(T) bool) []T {
	result := make([]T, 0)
	for _, v := range collection {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// 映射集合
func Map[T any, R any](collection []T, transform func(T) R) []R {
	result := make([]R, 0)
	for _, v := range collection {
		result = append(result, transform(v))
	}
	return result
}

// 查找集合中的第一个匹配项
func Find[T any](collection []T, predicate func(T) bool) (T, bool) {
	for _, v := range collection {
		if predicate(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// 检查集合是否为空
func IsEmptyCollection[T any](collection []T) bool {
	return len(collection) == 0
}

// 合并两个集合
func MergeCollections[T any](collection1, collection2 []T) []T {
	return append(collection1, collection2...)
}
