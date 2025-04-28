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
	options WebHostOptions
}

type WebHostOptions struct {
	Server ServerOptions
	Gin    GinOptions // gin配置
}

type ServerOptions struct {
	Port string `mapstructure:"port"`
}
type GinOptions struct {
	Mode string `mapstructure:"mode"`
}

const (
	port = "8080"
)

func NewWebHostBuilder() *WebHostBuilder {
	return &WebHostBuilder{
		ApplicationHostBuilder: NewApplicationHostBuilder(),
		options: WebHostOptions{
			Server: ServerOptions{
				Port: port,
			},
			Gin: GinOptions{
				Mode: gin.ReleaseMode,
			},
		},
	}
}

// 配置web服务器
func (b *WebHostBuilder) ConfigureWebServer(options WebHostOptions) *WebHostBuilder {

	if str.IsEmptyOrWhiteSpace(options.Server.Port) {
		panic("http server port is empty")
	}
	if str.IsEmptyOrWhiteSpace(options.Gin.Mode) {
		panic("http gin mode is empty")
	}

	b.options = options
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
	if err := host.Config().Unmarshal(&b.options); err != nil {
		return nil, fmt.Errorf("failed to bind config to WebHostOptions: %w", err)
	}

	return newWebApplication(WebApplicationOptions{
		Host:           host,
		WebHostOptions: b.options,
	}), nil
}
