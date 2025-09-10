package workit

import (
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// WebApplicationBuilder 构建web应用
type WebApplicationBuilder struct {
	*ApplicationBuilder
	auth       *AuthenticationBuilder
	author     *AuthorizationBuilder
	Config     *viper.Viper
	Logger     *zap.Logger
	Container  []fx.Option
	authopts   *AuthenticationOptions
	authoropts *AuthorizationOptions
}

// NewWebAppBuilder 创建WebApplicationBuilder
func NewWebAppBuilder() *WebApplicationBuilder {

	hostBuild := NewAppBuilder()

	return &WebApplicationBuilder{
		ApplicationBuilder: hostBuild,
		Config:             hostBuild.config,
	}
}

// AddAuthentication 添加鉴权方案
func (b *WebApplicationBuilder) AddAuthentication(options func(*AuthenticationOptions)) *AuthenticationBuilder {

	if b.authopts == nil {

		b.authopts = newAuthenticationOptions()
	}

	options(b.authopts)

	if b.authopts.DefaultScheme == "" {
		panic("default scheme is required")
	}

	b.AddServices(fx.Provide(func() *AuthenticationOptions { return b.authopts }))

	b.auth = newAuthenticationBuilder()

	return b.auth
}

// AddAuthorization 添加授权策略
func (b *WebApplicationBuilder) AddAuthorization(fn func(*AuthorizationOptions)) *AuthorizationBuilder {

	if b.authoropts == nil {

		b.authoropts = newAuthorizationOptions()
	}

	fn(b.authoropts)

	b.AddServices(fx.Provide(func() *AuthorizationOptions { return b.authoropts }))

	b.author = newAuthorizationBuilder()

	return b.author
}

// AddRouter 添加路由配置
func (b *WebApplicationBuilder) AddRouter(fn func(*RouterOptions)) *WebApplicationBuilder {

	if b.authopts == nil {
		b.authopts = newAuthenticationOptions()
	}

	if b.authoropts == nil {
		b.authoropts = newAuthorizationOptions()
	}

	opts := newRouterOptions(b.authopts, b.authoropts)

	fn(opts)

	return b
}

// AddDbContext 添加数据库配置
func (b *WebApplicationBuilder) AddDbContext(fn func(*DbContextOptions)) *WebApplicationBuilder {

	opts := newDatabaseOptions()

	fn(opts)

	b.Container = append(b.Container, opts.container...)

	return b
}

// AddCacheContext 添加缓存配置
func (b *WebApplicationBuilder) AddCacheContext(fn func(*CacheContextOptions)) *WebApplicationBuilder {

	opts := newCacheContextOptions()

	fn(opts)

	b.Container = append(b.Container, opts.container...)

	return b
}

// Build 构建应用
func (b *WebApplicationBuilder) Build(fn ...func(b *WebApplicationBuilder) WebApplication) WebApplication {

	// 1. 构建应用主机
	host := b.ApplicationBuilder.Build()

	// 2. 构建鉴权提供者
	if b.auth != nil {
		authProvider := b.auth.Build()
		host.container = append(host.container, fx.Supply(authProvider))
	} else {
		// 鉴权授权跳过用的同一个跳过配置,没有配置授权会报错
		host.container = append(host.container, fx.Supply(newAuthenticationOptions()))
		host.container = append(host.container, fx.Supply(newAuthenticateProvider(make(map[string]AuthenticationHandler))))
	}

	// 3. 构建授权提供者
	if b.author == nil {
		b.author = newAuthorizationBuilder()
		host.container = append(host.container, fx.Supply(newAuthorizationOptions()))
	}

	authorProvider := b.author.Build()
	host.container = append(host.container, fx.Supply(authorProvider))

	b.Container = append(b.Container, host.container...)
	b.Logger = host.logger

	// 4. 构建应用
	if len(fn) > 0 {
		return fn[0](b)
	}

	return newGinWebApplication(WebApplicationOptions{
		Config:    b.Config,
		Logger:    b.Logger,
		Container: b.Container,
	})
}
