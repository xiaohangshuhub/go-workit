package host

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/lxhanghub/newb/pkg/tools/str"
)

type Middleware interface {
	Handle() gin.HandlerFunc
	ShouldSkip(path string) bool
}

type WebHostBuilder struct {
	*ApplicationHostBuilder
	Server ServerOptions
}

type ServerOptions struct {
	Port string `mapstructure:"port"`
}

const (
	port = "8080"
)

func NewWebHostBuilder() *WebHostBuilder {

	hostBuild := NewApplicationHostBuilder()

	// 设置默认的web服务器端口
	hostBuild.config.SetDefault("server.port", port)

	return &WebHostBuilder{
		ApplicationHostBuilder: hostBuild,
		Server: ServerOptions{
			Port: port,
		},
	}
}

// 配置web服务器
func (b *WebHostBuilder) ConfigureWebServer(options ServerOptions) *WebHostBuilder {

	if str.IsEmptyOrWhiteSpace(options.Port) {
		panic("http server port is empty")
	}

	b.Server.Port = options.Port

	return b
}

// 构建应用
func (b *WebHostBuilder) Build() (*WebApplication, error) {

	// 1. 构建应用主机
	host, err := b.BuildHost()

	if err != nil {
		return nil, err
	}

	// 2. 绑定配置
	if err := host.Config().UnmarshalKey("server", &b.Server); err != nil {
		return nil, fmt.Errorf("failed to bind config to WebHostOptions: %w", err)
	}

	return newWebApplication(WebApplicationOptions{
		Host:   host,
		Server: b.Server,
	}), nil
}
