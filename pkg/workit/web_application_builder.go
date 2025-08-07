package workit

import (
	"fmt"
	"strings"

	"go.uber.org/fx"
)

type WebApplicationBuilder struct {
	*ApplicationBuilder
	*AuthenticationBuilder
	Server ServerOptions
}

type ServerOptions struct {
	Port     string `mapstructure:"port"`
	GrpcPort string `mapstructure:"grpc_port"`
}

const (
	port      = "8080"
	grpc_port = "50051"
)

func NewWebAppBuilder() *WebApplicationBuilder {

	hostBuild := NewAppBuilder()

	// 设置默认的web服务器端口
	hostBuild.config.SetDefault("server.port", port)

	return &WebApplicationBuilder{
		ApplicationBuilder: hostBuild,
		Server: ServerOptions{
			Port:     port,
			GrpcPort: grpc_port,
		},
	}
}

// 配置web服务器
func (b *WebApplicationBuilder) ConfigureWebServer(options ServerOptions) *WebApplicationBuilder {

	if strings.TrimSpace(options.Port) == "" {
		panic("http server port is empty")
	}

	b.Server.Port = options.Port

	return b
}

// 添加鉴权
func (b *WebApplicationBuilder) AddAuthentication(skip ...string) *AuthenticationBuilder {

	b.AddServices(fx.Provide(func() *AuthMiddlewareOptions {

		return &AuthMiddlewareOptions{
			SkipPaths: skip,
		}
	}))

	b.AuthenticationBuilder = NewAuthenticationBuilder()

	return b.AuthenticationBuilder
}

// 构建应用
func (b *WebApplicationBuilder) Build() (*WebApplication, error) {

	// 1. 构建应用主机
	host, err := b.ApplicationBuilder.Build()

	if err != nil {
		return nil, err
	}

	// 2. 绑定配置
	if err := host.Config().UnmarshalKey("server", &b.Server); err != nil {
		return nil, fmt.Errorf("failed to bind config to WebHostOptions: %w", err)
	}

	if b.AuthenticationBuilder != nil {
		// 3. 构建鉴权提供者
		authProvider := b.AuthenticationBuilder.Build()
		host.appoptions = append(host.appoptions, fx.Supply(authProvider))
	}

	return newWebApplication(WebApplicationOptions{
		Host:   host,
		Server: b.Server,
	}), nil
}
