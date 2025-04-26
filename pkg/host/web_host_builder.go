package host

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	"go.uber.org/fx"
)

type Middleware interface {
	Handle() gin.HandlerFunc
	ShouldSkip(path string) bool
}

type route struct {
	method  string
	path    string
	handler gin.HandlerFunc
}

type WebHostBuilder struct {
	*ApplicationHostBuilder
	engine      *gin.Engine
	middlewares []Middleware
	routes      []route
}

func NewWebHostBuilder() *WebHostBuilder {
	return &WebHostBuilder{
		ApplicationHostBuilder: NewApplicationHostBuilder(),
		engine:                 gin.New(),
		middlewares:            make([]Middleware, 0),
		routes:                 make([]route, 0),
	}
}

// 配置配置
func (b *WebHostBuilder) ConfigureAppConfiguration(fn func(ConfigBuilder)) *WebHostBuilder {
	b.ApplicationHostBuilder.ConfigureAppConfiguration(fn)
	return b
}

// 注册依赖
func (b *WebHostBuilder) ConfigureServices(opts ...fx.Option) *WebHostBuilder {
	b.ApplicationHostBuilder.ConfigureServices(opts...)
	return b
}

// 注册后台任务
func (b *WebHostBuilder) AddBackgroundService(ctor interface{}) *WebHostBuilder {
	b.ApplicationHostBuilder.AddBackgroundService(ctor)
	return b
}

// 注册中间件
func (b *WebHostBuilder) UseMiddleware(mws ...Middleware) *WebHostBuilder {
	b.middlewares = append(b.middlewares, mws...)
	return b
}

// 注册路由
func (b *WebHostBuilder) MapGet(path string, handler gin.HandlerFunc) *WebHostBuilder {
	b.routes = append(b.routes, route{"GET", path, handler})
	return b
}

func (b *WebHostBuilder) MapPost(path string, handler gin.HandlerFunc) *WebHostBuilder {
	b.routes = append(b.routes, route{"POST", path, handler})
	return b
}

func (b *WebHostBuilder) MapPut(path string, handler gin.HandlerFunc) *WebHostBuilder {
	b.routes = append(b.routes, route{"PUT", path, handler})
	return b
}

func (b *WebHostBuilder) MapDelete(path string, handler gin.HandlerFunc) *WebHostBuilder {
	b.routes = append(b.routes, route{"DELETE", path, handler})
	return b
}

// UseSwagger 配置Swagger文档
func (b *WebHostBuilder) UseSwagger() *WebHostBuilder {
	b.MapGet("/swagger/*any", gin.WrapH(http.HandlerFunc(swaggerFiles.Handler.ServeHTTP)))
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
	b.MapGet("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	return b
}

// UseGroup 配置路由组
func (b *WebHostBuilder) UseGroup(path string, fn func(r *RouterGroup)) *WebHostBuilder {
	group := &RouterGroup{
		path:    path,
		builder: b,
	}
	fn(group)
	return b
}

// RouterGroup 路由组
type RouterGroup struct {
	path    string
	builder *WebHostBuilder
}

func (g *RouterGroup) MapGet(path string, handler gin.HandlerFunc) {
	g.builder.MapGet(g.path+path, handler)
}

func (g *RouterGroup) MapPost(path string, handler gin.HandlerFunc) {
	g.builder.MapPost(g.path+path, handler)
}

func (g *RouterGroup) MapPut(path string, handler gin.HandlerFunc) {
	g.builder.MapPut(g.path+path, handler)
}

func (g *RouterGroup) MapDelete(path string, handler gin.HandlerFunc) {
	g.builder.MapDelete(g.path+path, handler)
}

// 构建应用
func (b *WebHostBuilder) Build() (*WebApplication, error) {
	for _, mw := range b.middlewares {
		handler := mw.Handle()
		b.engine.Use(func(c *gin.Context) {
			if !mw.ShouldSkip(c.Request.URL.Path) {
				handler(c)
			}
			c.Next()
		})
	}

	for _, r := range b.routes {
		switch r.method {
		case "GET":
			b.engine.GET(r.path, r.handler)
		case "POST":
			b.engine.POST(r.path, r.handler)
		case "PUT":
			b.engine.PUT(r.path, r.handler)
		case "DELETE":
			b.engine.DELETE(r.path, r.handler)
		}
	}

	host, err := b.BuildHost()
	if err != nil {
		return nil, err
	}

	return newWebApplication(host, b.engine), nil
}
