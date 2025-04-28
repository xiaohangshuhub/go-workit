package host

import (
	"fmt"
	"net/http"

	stdstrings "strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/lxhanghub/newb/pkg/tools/str"
	swaggerFiles "github.com/swaggo/files"
)

type Middleware interface {
	Handle() gin.HandlerFunc
	ShouldSkip(path string) bool
}

type route struct {
	method  string
	path    string
	handler any // 支持 gin.HandlerFunc 或 自定义 constructor
}

type WebHostBuilder struct {
	*ApplicationHostBuilder
	engine      *gin.Engine
	middlewares []Middleware
	routes      []route
	options     WebHostOptions
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
		middlewares:            make([]Middleware, 0),
		routes:                 make([]route, 0),
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

// 注册中间件
func (b *WebHostBuilder) UseMiddleware(mws ...Middleware) *WebHostBuilder {
	b.middlewares = append(b.middlewares, mws...)
	return b
}

// UseSwagger 配置Swagger文档
func (b *WebHostBuilder) UseSwagger() *WebHostBuilder {
	b.engine.GET("/swagger/*any", gin.WrapH(http.HandlerFunc(swaggerFiles.Handler.ServeHTTP)))
	return b
}

// UseCORS 配置跨域
func (b *WebHostBuilder) UseCORS() *WebHostBuilder {
	b.engine.Use(cors.Default())
	return b
}

// UseStaticFiles 配置静态文件服务
func (b *WebHostBuilder) UseStaticFiles(urlPath, root string) *WebHostBuilder {
	b.engine.Static(urlPath, root)
	return b
}

// UseHealthCheck 配置健康检查
func (b *WebHostBuilder) UseHealthCheck() *WebHostBuilder {
	b.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
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

	if str.IsEmptyOrWhiteSpace(b.options.Gin.Mode) {
		b.options.Gin.Mode = gin.ReleaseMode
	}

	switch stdstrings.ToLower(b.options.Gin.Mode) {
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}

	b.engine = gin.New()
	// 🔥 挂载自己的 zap logger + recovery
	b.engine.Use(NewGinZapLogger(b.logger))
	b.engine.Use(RecoveryWithZap(b.logger))

	if str.IsEmptyOrWhiteSpace(b.options.Server.Port) {
		b.options.Server.Port = port
	}

	for _, mw := range b.middlewares {
		// 创建一个局部变量，避免闭包捕获问题
		currentMiddleware := mw
		b.engine.Use(func(c *gin.Context) {
			if !currentMiddleware.ShouldSkip(c.Request.URL.Path) {
				handler := currentMiddleware.Handle()
				handler(c)
			} else {
				c.Next()
			}
		})
	}

	return newWebApplication(WebApplicationOptions{
		Host:    host,
		Handler: b.engine,
		Port:    b.options.Server.Port,
	}), nil
}
