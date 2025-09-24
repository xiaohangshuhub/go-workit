package webapp

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/xiaohangshuhub/go-workit/pkg/app"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/auth"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/authz"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/cachectx"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/dbctx"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/gin"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/localiza"

	"github.com/xiaohangshuhub/go-workit/pkg/webapp/ratelimit"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/router"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// WebApplicationBuilder 构建web应用
type WebApplicationBuilder struct {
	*app.Application
	*app.ApplicationBuilder
	Container     []fx.Option
	Config        *viper.Viper
	Logger        *zap.Logger
	authOpts      *auth.Options
	authzOpts     *authz.Options
	routeOpts     *router.Options
	localizaOpts  *localiza.Options
	rateLimitOpts *ratelimit.Options
	Router        *router.Router
}

// NewWebAppBuilder 创建WebApplicationBuilder
func NewBuilder() *WebApplicationBuilder {

	hostBuild := app.NewBuilder()

	return &WebApplicationBuilder{
		ApplicationBuilder: hostBuild,
		Config:             hostBuild.Config(),
	}
}

// AddAuthentication 添加鉴权方案
func (b *WebApplicationBuilder) AddAuthentication(options func(*auth.Options)) *WebApplicationBuilder {

	if b.authOpts == nil {

		b.authOpts = auth.NewOptions()
	}

	options(b.authOpts)

	if b.authOpts.DefaultScheme == "" {
		panic("default scheme is required")
	}

	return b
}

// AddAuthorization 添加授权策略
func (b *WebApplicationBuilder) AddAuthorization(fn func(*authz.Options)) *WebApplicationBuilder {

	if b.authzOpts == nil {

		b.authzOpts = authz.NewOptions()
	}

	fn(b.authzOpts)

	return b
}

// AddRouter 添加路由配置
func (b *WebApplicationBuilder) AddRouter(fn func(*router.Options)) *WebApplicationBuilder {

	opts := router.NewOptions()

	fn(opts)

	b.routeOpts = opts

	return b
}

// AddDbContext 添加数据库配置
func (b *WebApplicationBuilder) AddDbContext(fn func(*dbctx.Options)) *WebApplicationBuilder {

	opts := dbctx.NewOptions()

	fn(opts)

	b.Container = append(b.Container, opts.Container()...)

	return b
}

// AddCacheContext 添加缓存配置
func (b *WebApplicationBuilder) AddCacheContext(fn func(*cachectx.Options)) *WebApplicationBuilder {

	opts := cachectx.NewOptions()

	fn(opts)

	b.Container = append(b.Container, opts.Container()...)

	return b
}

func (b *WebApplicationBuilder) AddLocalization(fn func(*localiza.Options)) *WebApplicationBuilder {

	opts := localiza.NewOptions()

	fn(opts)

	b.localizaOpts = opts

	return b
}

// AddRateLimiter 添加限流配置
func (b *WebApplicationBuilder) AddRateLimiter(configure func(*ratelimit.Options)) *WebApplicationBuilder {
	opts := ratelimit.NewOptions()

	configure(opts)

	b.rateLimitOpts = opts

	return b
}

// Build 构建应用
func (b *WebApplicationBuilder) Build(fn ...func(b *WebApplicationBuilder) web.Application) web.Application {

	// 构建应用主机
	b.Application = b.ApplicationBuilder.Build()

	// 日志管理
	b.Logger = b.Application.Logger()

	// 构建路由
	if b.routeOpts == nil {
		b.routeOpts = router.NewOptions()
	}

	// 构建鉴权
	if b.authOpts == nil {
		b.authOpts = auth.NewOptions()
	}

	// 构建授权
	if b.authzOpts == nil {
		b.authzOpts = authz.NewOptions()
	}

	// 构建国际化
	if b.localizaOpts != nil {

		provider, err := localiza.NewBuilder(b.localizaOpts).Build()

		if err != nil {
			panic(fmt.Errorf("build localizer error: %w", err))
		}

		b.Application.AppendContainer(fx.Provide(
			func() web.Localization {
				return provider
			}))
	}

	// 构建限流
	if b.rateLimitOpts == nil {
		b.rateLimitOpts = ratelimit.NewOptions()
	}

	// 构建路由配置
	router := router.NewRouter(b.routeOpts, b.authOpts, b.authzOpts, b.rateLimitOpts)

	b.Router = router

	// 将路由配置注入容器供鉴权\授权\限流中间件使用
	b.AppendContainer(fx.Provide(func() web.Router {
		return b.Router
	}))

	b.Container = append(b.Container, b.Application.Container()...)

	// 构建应用
	if len(fn) > 0 {
		return fn[0](b)
	}

	return gin.NewGinWebApplication(web.InstanceConfig{
		Config:     b.Config,
		Logger:     b.Logger,
		Container:  b.Container,
		Applicaton: b.Application,
		Router:     b.Router,
	})
}
