package webapp

import (
	"fmt"

	"github.com/xiaohangshuhub/go-workit/pkg/app"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/auth"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/authz"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/cachectx"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/dbctx"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/ginx"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/localiza"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/reqdecp"

	"github.com/xiaohangshuhub/go-workit/pkg/webapp/ratelimit"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/router"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
	"go.uber.org/fx"
)

// WebApplicationBuilder 构建web应用
type WebApplicationBuilder struct {
	*app.ApplicationBuilder
	app           *app.Application
	authOpts      *auth.Options
	authzOpts     *authz.Options
	routeOpts     *router.Options
	localizaOpts  *localiza.Options
	rateLimitOpts *ratelimit.Options
	reqdecpOpts   *reqdecp.Options
	router        *router.Router
}

// NewWebAppBuilder 创建WebApplicationBuilder
func NewBuilder() *WebApplicationBuilder {

	hostBuild := app.NewBuilder()

	return &WebApplicationBuilder{
		ApplicationBuilder: hostBuild,
	}
}

// AddAuthentication 添加鉴权方案
func (b *WebApplicationBuilder) AddAuthentication(fn func(options *auth.Options)) *WebApplicationBuilder {

	if b.authOpts == nil {

		b.authOpts = auth.NewOptions()
	}

	fn(b.authOpts)

	if b.authOpts.DefaultScheme == "" {
		panic("default scheme is required")
	}

	return b
}

// AddAuthorization 添加授权策略
func (b *WebApplicationBuilder) AddAuthorization(fn func(options *authz.Options)) *WebApplicationBuilder {

	if b.authzOpts == nil {

		b.authzOpts = authz.NewOptions()
	}

	fn(b.authzOpts)

	return b
}

// AddRouter 添加路由配置
func (b *WebApplicationBuilder) AddRouter(fn func(options *router.Options)) *WebApplicationBuilder {

	opts := router.NewOptions()

	fn(opts)

	b.routeOpts = opts

	return b
}

// AddDbContext 添加数据库配置
func (b *WebApplicationBuilder) AddDbContext(fn func(options *dbctx.Options)) *WebApplicationBuilder {

	opts := dbctx.NewOptions()

	fn(opts)

	b.ApplicationBuilder.AddServices(opts.Container()...)

	return b
}

// AddCacheContext 添加缓存配置
func (b *WebApplicationBuilder) AddCacheContext(fn func(options *cachectx.Options)) *WebApplicationBuilder {

	opts := cachectx.NewOptions()

	fn(opts)

	b.ApplicationBuilder.AddServices(opts.Container()...)

	return b
}

func (b *WebApplicationBuilder) AddLocalization(fn func(options *localiza.Options)) *WebApplicationBuilder {

	opts := localiza.NewOptions()

	fn(opts)

	b.localizaOpts = opts

	return b
}

// AddRateLimiter 添加限流配置
func (b *WebApplicationBuilder) AddRateLimiter(fn func(options *ratelimit.Options)) *WebApplicationBuilder {
	opts := ratelimit.NewOptions()

	fn(opts)

	b.rateLimitOpts = opts

	return b
}

func (b *WebApplicationBuilder) AddRequestDecompression(fn ...func(options *reqdecp.Options)) *WebApplicationBuilder {

	opts := reqdecp.NewOptions()

	if len(fn) != 0 {
		fn[0](opts)
	}

	b.reqdecpOpts = opts

	return b
}

// Build 构建应用
func (b *WebApplicationBuilder) Build(fn ...func(b *WebApplicationBuilder) web.Application) web.Application {

	// 构建应用主机
	b.app = b.ApplicationBuilder.Build()

	if b.routeOpts == nil {
		b.routeOpts = router.NewOptions()
	}

	if b.authOpts == nil {
		b.authOpts = auth.NewOptions()
	}

	if b.authzOpts == nil {
		b.authzOpts = authz.NewOptions()
	}

	if b.rateLimitOpts == nil {
		b.rateLimitOpts = ratelimit.NewOptions()
	}

	if b.reqdecpOpts == nil {
		b.reqdecpOpts = reqdecp.NewOptions()
	}

	// 构建国际化
	if b.localizaOpts != nil {

		provider, err := localiza.NewBuilder(b.localizaOpts).Build()

		if err != nil {
			panic(fmt.Errorf("build localizer error: %w", err))
		}

		b.app.AppendContainer(fx.Provide(
			func() web.Localization {
				return provider
			}))
	}

	// 构建请求解压
	reqDecompressor := reqdecp.NewReqDecompressor(b.reqdecpOpts.Decompressions())

	b.app.AppendContainer(fx.Provide(func() web.ReqDecompressor {
		return reqDecompressor
	}))

	// 构建路由配置
	b.router = router.NewRouter(b.routeOpts, b.authOpts, b.authzOpts, b.rateLimitOpts)

	// 将路由配置注入容器供鉴权\授权\限流中间件使用
	b.app.AppendContainer(fx.Provide(func() web.Router {
		return b.router
	}))

	// 构建应用
	if len(fn) > 0 {
		return fn[0](b)
	}

	return ginx.NewWebApplication(b.app, b.router)
}

func (b *WebApplicationBuilder) App() *app.Application {
	return b.app
}

func (b *WebApplicationBuilder) Router() *router.Router {
	return b.router
}
