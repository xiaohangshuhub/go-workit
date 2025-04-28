package str

import "strings"

// 字符串是否为空
func IsEmpty(str string) bool {
	if str == "" {
		return true
	}
	return false
}

// 字符串是否全为空白字符
func IsWhiteSpace(str string) bool {
	return IsEmpty(strings.TrimSpace(str))
}

// 字符串是否为nil
func IsNil(str *string) bool {
	if str == nil {
		return true
	}
	return false
}

// 字符串是否不为空
func IsNotEmpty(str string) bool {
	return !IsEmpty(str)
}

// 字符串是否为空或全为空白字符
func IsEmptyOrWhiteSpace(str string) bool {
	return IsEmpty(str) || IsWhiteSpace(str)
}

// 字符串是否包含子串
func Contains(str, substr string) bool {
	return strings.Contains(str, substr)
}

// 将字符串转换为大写
func ToUpper(str string) string {
	return strings.ToUpper(str)
}

// 将字符串转换为小写
func ToLower(str string) string {
	return strings.ToLower(str)
}

// 反转字符串
func Reverse(str string) string {
	runes := []rune(str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// 分割字符串
func Split(str, sep string) []string {
	return strings.Split(str, sep)
}

// 替换字符串中的子串
func Replace(str, old, new string, n int) string {
	return strings.Replace(str, old, new, n)
}

// 连接字符串
func Join(elements []string, sep string) string {
	return strings.Join(elements, sep)
}

// 去除字符串两端的空白字符
func Trim(str string) string {
	return strings.TrimSpace(str)
}

// 检查字符串是否以特定前缀开头
func StartsWith(str, prefix string) bool {
	return strings.HasPrefix(str, prefix)
}

// 检查字符串是否以特定后缀结尾
func EndsWith(str, suffix string) bool {
	return strings.HasSuffix(str, suffix)
}

// 重复字符串
func Repeat(str string, count int) string {
	return strings.Repeat(str, count)
}
