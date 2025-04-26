package encrypt

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
)

// MD5 计算字符串的MD5值
func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// SHA256 计算字符串的SHA256值
func SHA256(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
