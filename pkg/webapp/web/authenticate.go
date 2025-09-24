package web

import "net/http"

// Authenticate 鉴权处理器接口
type Authenticate interface {
	Type() string                                           // 返回鉴权类型
	Authenticate(r *http.Request) (*ClaimsPrincipal, error) // 鉴权
}
