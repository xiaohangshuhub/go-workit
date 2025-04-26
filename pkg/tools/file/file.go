package file

import (
	"os"
)

// Exists 检查文件或目录是否存在
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// CreateDir 创建目录
func CreateDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

// ReadFile 读取文件内容
func ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile 写入文件内容
func WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}
