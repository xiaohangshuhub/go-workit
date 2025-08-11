package workit

import (
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type WebApplicationBuilder struct {
	*ApplicationBuilder
	*AuthenticationBuilder
	*AuthorizationBuilder
	Config    *viper.Viper
	Logger    *zap.Logger
	Container []fx.Option
}

func NewWebAppBuilder() *WebApplicationBuilder {

	hostBuild := NewAppBuilder()

	return &WebApplicationBuilder{
		ApplicationBuilder: hostBuild,
	}
}

// 添加鉴权
func (b *WebApplicationBuilder) AddAuthentication(authenticate ...AuthenticateOptions) *AuthenticationBuilder {

	if len(authenticate) == 0 {
		authenticate = append(authenticate, AuthenticateOptions{
			SkipPaths: make([]string, 0),
		})
	}

	b.AddServices(fx.Provide(func() AuthenticateOptions { return authenticate[0] }))

	b.AuthenticationBuilder = newAuthenticationBuilder()

	return b.AuthenticationBuilder
}

// 添加鉴权
func (b *WebApplicationBuilder) AddAuthorization(authorize ...AuthorizeOptions) *AuthorizationBuilder {

	b.AuthorizationBuilder = newAuthorizationBuilder(authorize...)

	return b.AuthorizationBuilder
}

// 构建应用
func (b *WebApplicationBuilder) Build(fn ...func(b *WebApplicationBuilder) WebApplication) (WebApplication, error) {

	// 1. 构建应用主机
	host, err := b.ApplicationBuilder.Build()

	if err != nil {
		return nil, err
	}

	// 2. 构建鉴权提供者
	if b.AuthenticationBuilder != nil {
		authProvider := b.AuthenticationBuilder.Build()
		host.container = append(host.container, fx.Supply(authProvider))
	} else {
		// 鉴权授权跳过用的同一个跳过配置,没有配置授权会报错
		host.container = append(host.container, fx.Supply(AuthenticateOptions{
			SkipPaths: make([]string, 0),
		}))

		host.container = append(host.container, fx.Supply(newAuthenticateProvider(make(map[string]AuthenticationHandler))))
	}

	// 3. 构建授权提供者
	if b.AuthorizationBuilder == nil {
		b.AuthorizationBuilder = newAuthorizationBuilder()
	}

	authorProvider := b.AuthorizationBuilder.Build()
	host.container = append(host.container, fx.Supply(authorProvider))

	b.Container = host.container
	b.Logger = host.logger
	b.Config = host.config

	// 4. 构建应用
	if len(fn) > 0 {
		return fn[0](b), nil
	}

	return newGinWebApplication(WebApplicationOptions{
		Config:    b.Config,
		Logger:    b.Logger,
		container: b.Container,
	}), nil
}
