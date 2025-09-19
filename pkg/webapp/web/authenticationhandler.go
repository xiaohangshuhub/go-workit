package web

import "net/http"

// AuthenticationHandler 鉴权处理器接口
type AuthenticationHandler interface {
	AuthType() string                                       // 返回鉴权类型
	Authenticate(r *http.Request) (*ClaimsPrincipal, error) // 鉴权
}
