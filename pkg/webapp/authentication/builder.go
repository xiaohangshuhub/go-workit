package authentication

import (
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/authentication/scheme/cookie"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/authentication/scheme/jwt"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
)

// AuthenticationBuilder 鉴权构建器
type Builder struct {
	schemes map[string]web.AuthenticationHandler
	*Options
}

// NewAuthenticationBuilder 新建鉴权构建器
func NewBuilder(options *Options) *Builder {
	return &Builder{
		Options: options,
		schemes: make(map[string]web.AuthenticationHandler),
	}
}

// AddScheme 注册新的鉴权方案
func (b *Builder) AddScheme(schemeName string, handler web.AuthenticationHandler) *Builder {

	if _, ok := b.schemes[schemeName]; ok {
		panic("scheme already exists:" + schemeName)
	}

	b.schemes[schemeName] = handler
	return b
}

// Schemes  返回所有注册的鉴权方案
func (b *Builder) Schemes() map[string]web.AuthenticationHandler {
	return b.schemes
}

// AddJwtBearer  注册新的 schemename JWT Bearer鉴权方案
func (b *Builder) AddJwtBearer(schemeName string, fn func(*jwt.Options)) *Builder {

	options := jwt.NewJwtBearerOptions()

	fn(options)

	b.AddScheme(schemeName, jwt.NewJWTBearerHandler(options))

	return b
}

// AddCookie  注册新的 schemename Cookie鉴权方案
func (b *Builder) AddCookie(schemeName string, fn func(*cookie.CookieOptions)) *Builder {

	options := cookie.NewCookieOptions()

	fn(options)

	b.AddScheme(schemeName, cookie.NewCookieAuthentication(options))

	return b
}

// Build  构建鉴权提供者
func (b *Builder) Build() (*Provider, error) {
	return newProvider(b.DefaultScheme, b.routeSchemesMap, b.allowAnonymous, b.schemes)
}
