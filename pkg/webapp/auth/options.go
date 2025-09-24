package auth

import (
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/auth/scheme/jwt"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
)

// Options 表示授权选项配置。
type Options struct {
	DefaultScheme string
	schemes       map[string]web.Authenticate
}

// NewOptions 创建一个新的 Options 实例
func NewOptions() *Options {
	opt := &Options{
		schemes: make(map[string]web.Authenticate),
	}

	return opt
}

// AddScheme 注册新的鉴权方案
func (o *Options) AddScheme(schemeName string, handler web.Authenticate) *Options {

	if _, ok := o.schemes[schemeName]; ok {
		panic("scheme already exists:" + schemeName)
	}

	o.schemes[schemeName] = handler
	return o
}

// Schemes  返回所有注册的鉴权方案
func (o *Options) Schemes() map[string]web.Authenticate {
	return o.schemes
}

// AddJwtBearer  注册新的 schemename JWT Bearer鉴权方案
func (o *Options) AddJwtBearer(schemeName string, fn func(*jwt.Options)) *Options {

	options := jwt.NewJwtBearerOptions()

	fn(options)

	o.AddScheme(schemeName, jwt.NewJWTBearerHandler(options))

	return o
}
