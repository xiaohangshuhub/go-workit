package webapp

import (
	"strings"

	"github.com/spf13/viper"
	"github.com/xiaohangshuhub/go-workit/pkg/app"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// WebApplicationBuilder 构建web应用
type WebApplicationBuilder struct {
	*app.ApplicationBuilder
	*app.Application
	*RouterOptions
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
func (b *WebApplicationBuilder) AddAuthentication(options func(*AuthenticationOptions)) *WebApplicationBuilder {

	if b.authopts == nil {

		b.authopts = newAuthenticationOptions()
	}

	options(b.authopts)

	if b.authopts.DefaultScheme == "" {
		panic("default scheme is required")
	}

	b.AddServices(fx.Provide(func() *AuthenticationOptions { return b.authopts }))

	b.authenticationBuilder = b.authOpts.AuthenticationBuilder

	return b
}

// AddAuthorization 添加授权策略
func (b *WebApplicationBuilder) AddAuthorization(fn func(*AuthorizationOptions)) *WebApplicationBuilder {

	if b.authoropts == nil {

		b.authoropts = newAuthorizationOptions()
	}

	fn(b.authoropts)

	b.AddServices(fx.Provide(func() *AuthorizationOptions { return b.authoropts }))

	b.authorizationBuilder = b.authorOpts.AuthorizationBuilder

	return b
}

// AddRouter 添加路由配置
func (b *WebApplicationBuilder) AddRouter(fn func(*RouterOptions)) *WebApplicationBuilder {

	if b.authopts == nil {
		b.authopts = newAuthenticationOptions()
	}

	if b.authoropts == nil {
		b.authoropts = newAuthorizationOptions()
	}

	if b.rateLimiterOptions == nil {
		b.rateLimiterOptions = newRateLimitOptions()
	}

	opts := newRouterOptions(b.authopts, b.authoropts, b.rateLimiterOptions)

	fn(opts)

	b.RouterOptions = opts

	b.AddServices(fx.Provide(func() *RouterOptions { return b.RouterOptions }))

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

	// 5. 路由
	if b.RouterOptions == nil {
		b.AddRouter(func(ro *RouterOptions) {})
	}

	for _, route := range b.routeConfigs {
		r := Route{Path: JoinPaths(route.Path), Methods: []RequestMethod{route.Method}}

		if route.AllowAnonymous {
			b.authOpts.useAllowAnonymous(r)
		}

		if len(route.Schemes) > 0 {
			b.authOpts.useRouteSchemes(RouteAuthenticationSchemes{Routes: []Route{r}, Schemes: route.Schemes})
		}

		if len(route.Policies) > 0 {
			b.authorOpts.useRoutePolicies(RouteAuthorizePolicies{Routes: []Route{r}, Policies: route.Policies})
		}

		if len(route.RateLimiter) > 0 {
			b.rateLimiterOptions.useRouteRateLimitPolicies(RouteRateLimitPolicies{Routes: []Route{r}, RateLimitPolicy: route.RateLimiter})
		}

	}
	for _, group := range b.groupConfigs {

		for _, route := range group.Routes {
			r := Route{Path: JoinPaths(group.Prefix, route.Path), Methods: []RequestMethod{route.Method}}

			if route.AllowAnonymous || group.AllowAnonymous {
				b.authOpts.useAllowAnonymous(r)
			}

			if len(route.Schemes) > 0 || len(group.Schemes) > 0 {

				schems := []string{}
				schems = append(schems, route.Schemes...)
				schems = append(schems, group.Schemes...)

				b.authOpts.useRouteSchemes(RouteAuthenticationSchemes{Routes: []Route{r}, Schemes: schems})
			}

			if len(route.Policies) > 0 {

				policies := []string{}
				policies = append(policies, route.Policies...)
				policies = append(policies, group.Policies...)

				b.authorOpts.useRoutePolicies(RouteAuthorizePolicies{Routes: []Route{r}, Policies: policies})
			}

			if len(route.RateLimiter) > 0 || len(group.RateLimiter) > 0 {

				rateLimiter := []string{}
				rateLimiter = append(rateLimiter, route.RateLimiter...)
				rateLimiter = append(rateLimiter, group.RateLimiter...)

				b.rateLimiterOptions.useRouteRateLimitPolicies(RouteRateLimitPolicies{Routes: []Route{r}, RateLimitPolicy: rateLimiter})
			}

		}

	}

	// 构建应用
	if len(fn) > 0 {
		return fn[0](b)
	}

	return newGinWebApplication(WebApplicationOptions{
		Config:        b.Config,
		Logger:        b.Logger,
		Container:     b.Container,
		App:           b.Application,
		RouterOptions: b.RouterOptions,
	})
}

// JoinPaths 拼接多个 URL path 段，自动处理斜杠问题。
// 例如：
//
//	JoinPaths("/api", "v1", "users")      => "/api/v1/users"
//	JoinPaths("/api/", "/v1/", "/users/") => "/api/v1/users/"
//	JoinPaths("/", "/")                   => "/"
func JoinPaths(paths ...string) string {
	if len(paths) == 0 {
		return "/"
	}

	segments := []string{}
	for i, p := range paths {
		if p == "" {
			continue
		}
		if i == 0 {
			// 保留第一个的前导斜杠
			p = strings.TrimRight(p, "/")
		} else if i == len(paths)-1 {
			// 最后一个保留尾部斜杠（如果有）
			hasSlash := strings.HasSuffix(p, "/")
			p = strings.Trim(p, "/")
			if hasSlash {
				p += "/"
			}
		} else {
			// 中间部分，去掉首尾斜杠
			p = strings.Trim(p, "/")
		}
		if p != "" {
			segments = append(segments, p)
		}
	}

	result := strings.Join(segments, "/")

	// 确保以 / 开头
	if !strings.HasPrefix(result, "/") {
		result = "/" + result
	}
	return result
}
