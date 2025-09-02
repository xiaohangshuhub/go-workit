package workit

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// CookieAuthenticationHandler Cookie 认证处理器
type CookieAuthenticationHandler struct {
	options       *CookieOptions
	dataProtector DataProtector
}

// DataProtector 数据保护接口，用于加密和解密认证票据
type DataProtector interface {
	Protect(data []byte) ([]byte, error)
	Unprotect(protectedData []byte) ([]byte, error)
}

// AESDataProtector 使用 AES 加密实现数据保护
type AESDataProtector struct {
	key []byte
}

// AuthenticationTicket 认证票据，包含用户声明和认证属性
type AuthenticationTicket struct {
	ClaimsPrincipal ClaimsPrincipal `json:"claimsPrincipal"`
	Properties      AuthProperties  `json:"properties"`
}

// AuthProperties 认证属性
type AuthProperties struct {
	IssuedUtc    time.Time `json:"issuedUtc"`
	ExpiresUtc   time.Time `json:"expiresUtc"`
	IsPersistent bool      `json:"isPersistent"`
	RedirectUri  string    `json:"redirectUri,omitempty"`
}

// NewCookieAuthentication 创建新的 Cookie 认证处理器
func newCookieAuthentication(options *CookieOptions) *CookieAuthenticationHandler {
	if options.DataProtectionKey == "" {
		panic("DataProtectionKey must be set")
	}

	// 创建数据保护器
	protector, err := newAESDataProtector(options.DataProtectionKey)
	if err != nil {
		panic(err)
	}

	return &CookieAuthenticationHandler{
		options:       options,
		dataProtector: protector,
	}
}

// Scheme 认证方案名称
func (h *CookieAuthenticationHandler) Scheme() string {
	return "Cookie"
}

// Authenticate 从请求中认证 Cookie
func (a *CookieAuthenticationHandler) Authenticate(r *http.Request) (*ClaimsPrincipal, error) {
	cookie, err := r.Cookie(a.options.Name)
	if err != nil {
		return nil, errors.New("cookie not found")
	}

	// 解码 Base64
	decoded, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil, errors.New("invalid cookie format")
	}

	// 解密数据
	unprotected, err := a.dataProtector.Unprotect(decoded)
	if err != nil {
		return nil, fmt.Errorf("failed to unprotect cookie: %v", err)
	}

	// 反序列化认证票据
	var ticket AuthenticationTicket
	if err := json.Unmarshal(unprotected, &ticket); err != nil {
		return nil, fmt.Errorf("failed to deserialize ticket: %v", err)
	}

	// 检查过期时间
	if time.Now().UTC().After(ticket.Properties.ExpiresUtc) {
		return nil, errors.New("authentication ticket expired")
	}

	return &ticket.ClaimsPrincipal, nil
}

// NewAESDataProtector 创建 AES 数据保护器
func newAESDataProtector(key string) (*AESDataProtector, error) {
	// 确保密钥长度为 16, 24 或 32 字节
	keyBytes := []byte(key)
	if len(keyBytes) != 16 && len(keyBytes) != 24 && len(keyBytes) != 32 {
		return nil, errors.New("key must be 16, 24 or 32 bytes long")
	}

	return &AESDataProtector{key: keyBytes}, nil
}

// Protect 加密数据
func (p *AESDataProtector) Protect(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(p.key)
	if err != nil {
		return nil, err
	}

	// 创建 GCM 模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 生成随机 nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// 加密数据
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// Unprotect 解密数据
func (p *AESDataProtector) Unprotect(protectedData []byte) ([]byte, error) {
	block, err := aes.NewCipher(p.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(protectedData) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := protectedData[:nonceSize], protectedData[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
