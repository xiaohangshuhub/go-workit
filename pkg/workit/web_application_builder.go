package workit

import (
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// WebApplicationBuilder 构建web应用
type WebApplicationBuilder struct {
	*ApplicationBuilder
	*AuthenticationBuilder
	*AuthorizationBuilder
	Config    *viper.Viper
	Logger    *zap.Logger
	Container []fx.Option
}

// NewWebAppBuilder 创建WebApplicationBuilder
func NewWebAppBuilder() *WebApplicationBuilder {

	hostBuild := NewAppBuilder()

	return &WebApplicationBuilder{
		ApplicationBuilder: hostBuild,
	}
}

// AddAuthentication 添加鉴权
func (b *WebApplicationBuilder) AddAuthentication(options func(*AuthenticateOptions)) *AuthenticationBuilder {

	opts := newAuthenticateOptions()

	options(opts)

	if opts.DefaultScheme == "" {
		panic("default scheme is required")
	}

	b.AddServices(fx.Provide(func() *AuthenticateOptions { return opts }))

	b.AuthenticationBuilder = newAuthenticationBuilder()

	return b.AuthenticationBuilder
}

// AddAuthorization 添加鉴权
func (b *WebApplicationBuilder) AddAuthorization(fn func(*AuthorizeOptions)) *AuthorizationBuilder {

	opts := newAuthorizeOptions()

	fn(opts)

	b.AddServices(fx.Provide(func() *AuthorizeOptions { return opts }))

	b.AuthorizationBuilder = newAuthorizationBuilder()

	return b.AuthorizationBuilder
}

func (b *WebApplicationBuilder) AddRouter(fn func(*RouterOptions)) *WebApplicationBuilder {

	opts := newRouterOptions()

	fn(opts)

	return b
}

// Build 构建应用
func (b *WebApplicationBuilder) Build(fn ...func(b *WebApplicationBuilder) WebApplication) WebApplication {

	// 1. 构建应用主机
	host := b.ApplicationBuilder.Build()

	// 2. 构建鉴权提供者
	if b.AuthenticationBuilder != nil {
		authProvider := b.AuthenticationBuilder.Build()
		host.container = append(host.container, fx.Supply(authProvider))
	} else {
		// 鉴权授权跳过用的同一个跳过配置,没有配置授权会报错
		host.container = append(host.container, fx.Supply(newAuthenticateOptions()))
		host.container = append(host.container, fx.Supply(newAuthenticateProvider(make(map[string]AuthenticationHandler))))
	}

	// 3. 构建授权提供者
	if b.AuthorizationBuilder == nil {
		b.AuthorizationBuilder = newAuthorizationBuilder()
		host.container = append(host.container, fx.Supply(newAuthorizeOptions()))
	}

	authorProvider := b.AuthorizationBuilder.Build()
	host.container = append(host.container, fx.Supply(authorProvider))

	b.Container = host.container
	b.Logger = host.logger
	b.Config = host.config

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
