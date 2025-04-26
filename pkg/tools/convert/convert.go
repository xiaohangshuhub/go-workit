package convert

import (
	"encoding/json"
	"strconv"
)

// ToString 将任意类型转换为字符串
func ToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case bool:
		return strconv.FormatBool(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		b, _ := json.Marshal(value)
		return string(b)
	}
}

// ToInt 将字符串转换为整数
func ToInt(str string) (int, error) {
	return strconv.Atoi(str)
}

// ToBool 将字符串转换为布尔值
func ToBool(str string) (bool, error) {
	return strconv.ParseBool(str)
}
