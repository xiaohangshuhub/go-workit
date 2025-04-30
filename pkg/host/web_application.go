package host

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	stdstrings "strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Middleware interface {
	Handle() gin.HandlerFunc
	ShouldSkip(path string) bool
}

type Environment struct {
	IsDevelopment bool // 是否开发环境
}

type WebApplication struct {
	*Application
	handler            http.Handler
	server             *http.Server
	routeRegistrations []interface{}
	middlewares        []Middleware
	serverOptons       ServerOptions
	Env                Environment //环境
}

type WebApplicationOptions struct {
	Host   *Application
	Server ServerOptions
}

func newWebApplication(optinos WebApplicationOptions) *WebApplication {

	if optinos.Server == (ServerOptions{}) {
		panic("web host options is empty")
	}

	mode := optinos.Host.config.GetString("gin.mode")

	switch stdstrings.ToLower(mode) {
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}

	gin := gin.New()
	// 🔥 挂载自己的 zap logger + recovery
	gin.Use(NewGinZapLogger(optinos.Host.logger))

	gin.Use(RecoveryWithZap(optinos.Host.logger))

	return &WebApplication{
		Application:  optinos.Host,
		handler:      gin,
		middlewares:  make([]Middleware, 0),
		serverOptons: optinos.Server,
	}
}

func (app *WebApplication) Run(ctx ...context.Context) error {
	var appCtx context.Context
	var cancel context.CancelFunc

	// 如果调用者未传递上下文，则创建默认上下文
	if len(ctx) == 0 || ctx[0] == nil {
		appCtx, cancel = context.WithCancel(context.Background())
		defer cancel()

		// 捕获系统信号，优雅关闭
		go func() {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			<-sigChan
			fmt.Println("Received shutdown signal")
			cancel()
		}()
	} else {
		// 使用调用者传递的上下文
		appCtx = ctx[0]
	}

	app.server = &http.Server{
		Addr:         ":" + app.serverOptons.Port,
		Handler:      app.handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动 HTTP 服务器
	go func() {
		app.Logger().Info("HTTP server starting...", zap.String("port", app.serverOptons.Port))

		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.Logger().Error("HTTP server ListenAndServe error", zap.Error(err))
		}
	}()

	for _, mw := range app.middlewares {
		// 创建一个局部变量，避免闭包捕获问题
		currentMiddleware := mw
		app.engine().Use(func(c *gin.Context) {
			if !currentMiddleware.ShouldSkip(c.Request.URL.Path) {
				handler := currentMiddleware.Handle()
				handler(c)
			} else {
				c.Next()
			}
		})
	}

	for _, r := range app.routeRegistrations {
		app.appoptions = append(app.appoptions, fx.Invoke(r))
	}

	app.appoptions = append(app.appoptions,
		fx.Supply(app.handler.(*gin.Engine)),
	)

	app.app = fx.New(app.appoptions...)

	// 启动应用程序
	if err := app.Start(appCtx); err != nil {
		return fmt.Errorf("start host failed: %w", err)
	}

	// 等待上下文被取消
	<-appCtx.Done()

	// 优雅关闭服务器
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server failed: %w", err)
	}

	return app.Stop(shutdownCtx)
}

func (a *WebApplication) MapRoutes(registerFunc interface{}) *WebApplication {
	a.routeRegistrations = append(a.routeRegistrations, registerFunc)
	return a
}

// UseSwagger 配置Swagger
func (a *WebApplication) UseSwagger() *WebApplication {
	a.engine().GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return a
}

// UseCORS 配置跨域
func (a *WebApplication) UseCORS() *WebApplication {
	a.engine().Use(cors.Default())
	return a
}

// UseStaticFiles 配置静态文件
func (a *WebApplication) UseStaticFiles(urlPath, root string) *WebApplication {
	a.engine().Static(urlPath, root)
	return a
}

// UseHealthCheck 配置健康检查
func (a *WebApplication) UseHealthCheck() *WebApplication {
	a.engine().GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	return a
}

func (a *WebApplication) engine() *gin.Engine {
	return a.handler.(*gin.Engine)
}

// 注册中间件
func (b *WebApplication) UseMiddleware(mws ...Middleware) *WebApplication {
	b.middlewares = append(b.middlewares, mws...)
	return b
}
