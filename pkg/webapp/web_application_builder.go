package webapp

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"github.com/xiaohangshuhub/go-workit/pkg/app"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/authentication"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/authorization"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/gin"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/localization"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/ratelimit"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/router"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// WebApplicationBuilder 构建web应用
type WebApplicationBuilder struct {
	*app.ApplicationBuilder
	*app.Application
	Config                *viper.Viper
	Logger                *zap.Logger
	Container             []fx.Option
	routeOptions          *router.Options
	authenticationOptions *authentication.Options
	authorizationOptions  *authorization.Options
	localizerOptions      *localization.Options
	rateLimiterOptions    *ratelimit.Options
	routeProvider         *router.Provider
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
func (b *WebApplicationBuilder) AddAuthentication(options func(*authentication.Options)) *WebApplicationBuilder {

	if b.authenticationOptions == nil {

		b.authenticationOptions = authentication.NewOptions()
	}

	options(b.authenticationOptions)

	if b.authenticationOptions.DefaultScheme == "" {
		panic("default scheme is required")
	}

	return b
}

// AddAuthorization 添加授权策略
func (b *WebApplicationBuilder) AddAuthorization(fn func(*authorization.Options)) *WebApplicationBuilder {

	if b.authorizationOptions == nil {

		b.authorizationOptions = authorization.NewOptions()
	}

	fn(b.authorizationOptions)

	return b
}

// AddRouter 添加路由配置
func (b *WebApplicationBuilder) AddRouter(fn func(*router.Options)) *WebApplicationBuilder {

	if b.authenticationOptions == nil {
		b.authenticationOptions = authentication.NewOptions()
	}

	if b.authorizationOptions == nil {
		b.authorizationOptions = authorization.NewOptions()
	}

	if b.rateLimiterOptions == nil {
		b.rateLimiterOptions = ratelimit.NewOptions()
	}

	opts := router.NewOptions()

	fn(opts)

	b.routeOptions = opts

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

func (b *WebApplicationBuilder) AddLocalization(fn func(*localization.Options)) *WebApplicationBuilder {

	opts := localization.NewLocalizerOptions()

	fn(opts)

	b.localizerOptions = opts

	return b
}

// AddRateLimiter 添加限流配置
func (b *WebApplicationBuilder) AddRateLimiter(configure func(*ratelimit.Options)) *WebApplicationBuilder {
	opts := ratelimit.NewOptions()

	configure(opts)

	b.rateLimiterOptions = opts

	b.AddServices(fx.Provide(func() *ratelimit.Options { return b.rateLimiterOptions }))

	return b
}

// Build 构建应用
func (b *WebApplicationBuilder) Build(fn ...func(b *WebApplicationBuilder) web.Application) web.Application {

	// 1. 构建应用主机
	b.Application = b.ApplicationBuilder.Build()

	// 2. 构建鉴权提供者
	if b.authenticationOptions != nil {

		provider, err := b.authenticationOptions.Build()

		if err != nil {
			panic(fmt.Errorf("build authentication error: %w", err))
		}

		b.Application.AppendContainer(fx.Provide(
			func() web.AuthenticateProvider {
				return provider
			}))

	}

	// 3. 构建授权提供者
	if b.authorizationOptions != nil {
		provider, err := b.authorizationOptions.Build()

		if err != nil {
			panic(fmt.Errorf("build authorization error: %w", err))
		}

		b.Application.AppendContainer(fx.Provide(
			func() web.AuthorizationProvider {
				return provider
			}))
	}

	// 4. 构建国际化服务
	if b.localizerOptions != nil {

		provider, err := b.localizerOptions.Build()

		if err != nil {
			panic(fmt.Errorf("build localizer error: %w", err))
		}

		b.Application.AppendContainer(fx.Provide(
			func() web.LocalizationProvider {
				return provider
			}))
	}

	// 5. 路由
	if b.routeOptions != nil {

		provider := b.routeOptions.Build()

		for _, route := range provider.RouteConfig() {
			r := router.Route{Path: JoinPaths(route.Path), Methods: []router.RequestMethod{route.Method}}

			if route.AllowAnonymous {
				b.authenticationOptions.UseAllowAnonymous(r)
			}

			if len(route.Schemes) > 0 {
				b.authenticationOptions.UseRouteSchemes(authentication.RouteSchemes{Routes: []router.Route{r}, Schemes: route.Schemes})
			}

			if len(route.Policies) > 0 {
				b.authorizationOptions.UseRoutePolicies(authorization.RoutePolicies{Routes: []router.Route{r}, Policies: route.Policies})
			}

			if len(route.RateLimiter) > 0 {
				b.rateLimiterOptions.UseRouteRateLimitPolicies(ratelimit.RoutePolicies{Routes: []router.Route{r}, RateLimitPolicy: route.RateLimiter})
			}

		}
		for _, group := range provider.GroupRouteConfig() {

			for _, route := range group.Routes {
				r := router.Route{Path: JoinPaths(group.Prefix, route.Path), Methods: []router.RequestMethod{route.Method}}

				if route.AllowAnonymous || group.AllowAnonymous {
					b.authenticationOptions.UseAllowAnonymous(r)
				}

				if len(route.Schemes) > 0 || len(group.Schemes) > 0 {

					schems := []string{}
					schems = append(schems, route.Schemes...)
					schems = append(schems, group.Schemes...)

					b.authenticationOptions.UseRouteSchemes(authentication.RouteSchemes{Routes: []router.Route{r}, Schemes: schems})
				}

				if len(route.Policies) > 0 {

					policies := []string{}
					policies = append(policies, route.Policies...)
					policies = append(policies, group.Policies...)

					b.authorizationOptions.UseRoutePolicies(authorization.RoutePolicies{Routes: []router.Route{r}, Policies: policies})
				}

				if len(route.RateLimiter) > 0 || len(group.RateLimiter) > 0 {

					rateLimiter := []string{}
					rateLimiter = append(rateLimiter, route.RateLimiter...)
					rateLimiter = append(rateLimiter, group.RateLimiter...)

					b.rateLimiterOptions.UseRouteRateLimitPolicies(ratelimit.RoutePolicies{Routes: []router.Route{r}, RateLimitPolicy: rateLimiter})
				}

			}

		}

		b.routeProvider = provider
	}

	b.Container = append(b.Container, b.Application.Container()...)
	b.Logger = b.Application.Logger()

	// 构建应用
	if len(fn) > 0 {
		return fn[0](b)
	}

	return gin.NewGinWebApplication(web.InstanceConfig{
		Config:         b.Config,
		Logger:         b.Logger,
		Container:      b.Container,
		Applicaton:     b.Application,
		RouterProvider: b.routeProvider,
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
