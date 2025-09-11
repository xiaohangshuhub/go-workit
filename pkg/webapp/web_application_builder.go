package webapp

import (
	"github.com/spf13/viper"
	"github.com/xiaohangshuhub/go-workit/pkg/app"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// WebApplicationBuilder 构建web应用
type WebApplicationBuilder struct {
	*app.ApplicationBuilder
	*app.Application
	authenticationBuilder *AuthenticationBuilder
	authorizationBuilder  *AuthorizationBuilder
	Config                *viper.Viper
	Logger                *zap.Logger
	Container             []fx.Option
	authopts              *AuthenticationOptions
	authoropts            *AuthorizationOptions
	localizerBuilder      *LocalizerBuilder
	localizerOptions      *LocalizationOptions
	rateLimiterOptions    *RateLimitOptions
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
func (b *WebApplicationBuilder) AddAuthentication(options func(*AuthenticationOptions)) *AuthenticationBuilder {

	if b.authopts == nil {

		b.authopts = newAuthenticationOptions()
	}

	options(b.authopts)

	if b.authopts.DefaultScheme == "" {
		panic("default scheme is required")
	}

	b.AddServices(fx.Provide(func() *AuthenticationOptions { return b.authopts }))

	b.authenticationBuilder = newAuthenticationBuilder()

	return b.authenticationBuilder
}

// AddAuthorization 添加授权策略
func (b *WebApplicationBuilder) AddAuthorization(fn func(*AuthorizationOptions)) *AuthorizationBuilder {

	if b.authoropts == nil {

		b.authoropts = newAuthorizationOptions()
	}

	fn(b.authoropts)

	b.AddServices(fx.Provide(func() *AuthorizationOptions { return b.authoropts }))

	b.authorizationBuilder = newAuthorizationBuilder()

	return b.authorizationBuilder
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

func (b *WebApplicationBuilder) AddLocalization(fn func(*LocalizationOptions)) *WebApplicationBuilder {

	opts := newLocalizerOptions()

	fn(opts)

	b.localizerBuilder = newLocalizerBuilder(opts.DefaultLanguage, opts.SupportedLanguages, opts.TranslationsDir, opts.FileType)

	b.localizerOptions = opts

	b.AddServices(fx.Provide(func() *LocalizationOptions { return b.localizerOptions }))

	return b
}

// AddRateLimiter 添加限流配置
func (b *WebApplicationBuilder) AddRateLimiter(configure func(*RateLimitOptions)) *WebApplicationBuilder {
	opts := newRateLimitOptions()

	configure(opts)

	b.rateLimiterOptions = opts

	b.AddServices(fx.Provide(func() *RateLimitOptions { return b.rateLimiterOptions }))

	return b
}

// Build 构建应用
func (b *WebApplicationBuilder) Build(fn ...func(b *WebApplicationBuilder) WebApplication) WebApplication {

	// 1. 构建应用主机
	b.Application = b.ApplicationBuilder.Build()

	// 2. 构建鉴权提供者
	if b.authenticationBuilder != nil {
		authProvider := b.authenticationBuilder.Build()
		b.Application.AppendContainer(fx.Supply(authProvider))
	} else {
		// 鉴权授权跳过用的同一个跳过配置,没有配置授权会报错
		b.Application.AppendContainer(fx.Supply(newAuthenticationOptions()))
		b.Application.AppendContainer(fx.Supply(newAuthenticateProvider(make(map[string]AuthenticationHandler))))
	}

	// 3. 构建授权提供者
	if b.authorizationBuilder == nil {
		b.authorizationBuilder = newAuthorizationBuilder()
		b.Application.AppendContainer(fx.Supply(newAuthorizationOptions()))
	}

	authorProvider := b.authorizationBuilder.Build()
	b.Application.AppendContainer(fx.Supply(authorProvider))

	b.Container = append(b.Container, b.Application.Container()...)
	b.Logger = b.Application.Logger()

	// 4. 构建国际化服务
	if b.localizerBuilder != nil {
		bundle, err := b.localizerBuilder.Build()
		if err != nil {
			panic(err)
		}
		b.localizerOptions.Bundle = bundle
	}

	// 构建应用
	if len(fn) > 0 {
		return fn[0](b)
	}

	return newGinWebApplication(WebApplicationOptions{
		Config:    b.Config,
		Logger:    b.Logger,
		Container: b.Container,
		App:       b.Application,
	})
}
