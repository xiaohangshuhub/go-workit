package workit

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Middleware interface {
	Handle() echo.MiddlewareFunc
	ShouldSkip(path string) bool
}

type Environment struct {
	IsDevelopment bool // 是否开发环境
}

type WebApplication struct {
	*Application
	handler                 http.Handler
	server                  *http.Server
	routeRegistrations      []interface{}
	middlewares             []Middleware
	serverOptons            ServerOptions
	Env                     Environment //环境
	grpcServiceConstructors []interface{}
}

type WebApplicationOptions struct {
	Host   *Application
	Server ServerOptions
}

func newWebApplication(options WebApplicationOptions) *WebApplication {
	env := Environment{}

	if options.Server == (ServerOptions{}) {
		panic("web host options is empty")
	}

	e := echo.New()

	// 读取 echo.debug 配置
	debug := options.Host.config.GetBool("echo.debug")
	switch {
	case debug:
		// Debug模式
		env.IsDevelopment = true
		e.Debug = true
		e.HideBanner = false
		e.HidePort = false
		options.Host.logger.Info("Running in Debug mode")
	default:
		// Release模式
		e.Debug = false
		e.HideBanner = true
		e.HidePort = true
		options.Host.logger.Info("Running in Release mode")
	}

	// 替代 recovery 和 logger 使用 zap

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if e.Debug {
				// Debug模式，全部打日志
				options.Host.logger.Info("request",
					zap.String("URI", v.URI),
					zap.Int("status", v.Status),
				)
			} else {
				// Release模式，只打非200
				if v.Status != http.StatusOK {
					options.Host.logger.Info("request",
						zap.String("URI", v.URI),
						zap.Int("status", v.Status),
					)
				}
			}
			return nil
		},
	}))

	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			options.Host.logger.Error("panic recovered",
				zap.Error(err),
				zap.ByteString("stack", stack),
			)
			return nil
		},
	}))

	return &WebApplication{
		Application:  options.Host,
		handler:      e,
		middlewares:  make([]Middleware, 0),
		serverOptons: options.Server,
		Env:          env,
	}
}

func (app *WebApplication) Run(ctx ...context.Context) error {
	var appCtx context.Context
	var cancel context.CancelFunc

	if len(ctx) == 0 || ctx[0] == nil {
		appCtx, cancel = context.WithCancel(context.Background())
		defer cancel()

		go func() {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			<-sigChan
			fmt.Println("Received shutdown signal")
			cancel()
		}()
	} else {
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

	// 启动 gRPC 服务器
	if len(app.grpcServiceConstructors) > 0 {

		app.appoptions = append(app.appoptions,
			fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger) *grpc.Server {
				return NewGrpcServer(lc, logger, app.serverOptons)
			}),
		)
	}

	// 注册中间件（适配 interface）
	for _, mw := range app.middlewares {
		currentMiddleware := mw
		app.engine().Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				if !currentMiddleware.ShouldSkip(c.Path()) {
					return currentMiddleware.Handle()(next)(c)
				}
				return next(c)
			}
		})
	}

	// 注册路由
	for _, r := range app.routeRegistrations {
		app.appoptions = append(app.appoptions, fx.Invoke(r))
	}

	app.appoptions = append(app.appoptions,
		fx.Supply(app.handler.(*echo.Echo)), // echo.Echo 实现 http.Handler
	)

	app.app = fx.New(app.appoptions...)

	if err := app.Start(appCtx); err != nil {
		return fmt.Errorf("start host failed: %w", err)
	}

	<-appCtx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server failed: %w", err)
	}

	return app.Stop(shutdownCtx)
}

// UseStaticFiles 配置静态文件
func (a *WebApplication) UseStaticFiles(urlPath, root string) *WebApplication {
	a.engine().Static(urlPath, root)
	return a
}

// 健康检查
func (a *WebApplication) UseHealthCheck() *WebApplication {
	a.engine().GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	return a
}

// Swagger 支持
func (a *WebApplication) UseSwagger() *WebApplication {
	a.engine().GET("/swagger/*", echoSwagger.WrapHandler)
	return a
}

// CORS 支持
func (a *WebApplication) UseCORS(c ...middleware.CORSConfig) *WebApplication {
	if len(c) > 0 {
		a.engine().Use(middleware.CORSWithConfig(c[0]))
		return a
	}

	a.engine().Use(middleware.CORS())
	return a
}

// 路由注册
func (a *WebApplication) MapRoutes(registerFunc interface{}) *WebApplication {
	a.routeRegistrations = append(a.routeRegistrations, registerFunc)
	return a
}

// 中间件
func (a *WebApplication) UseMiddleware(mws ...Middleware) *WebApplication {
	a.middlewares = append(a.middlewares, mws...)
	return a
}
func (a *WebApplication) engine() *echo.Echo {
	return a.handler.(*echo.Echo)
}

func (app *WebApplication) MapGrpcServices(constructors ...interface{}) *WebApplication {
	for _, constructor := range constructors {
		app.grpcServiceConstructors = append(app.grpcServiceConstructors, constructor)
		app.appoptions = append(app.appoptions, fx.Provide(constructor))

		// 推断构造函数的返回类型
		constructorType := reflect.TypeOf(constructor)
		if constructorType.Kind() != reflect.Func || constructorType.NumOut() == 0 {
			panic("MapGrpcServices: constructor must be a function with at least one return value")
		}

		serviceType := constructorType.Out(0)

		// 对每个具体服务构造出一个 fx.Invoke
		invokeFn := makeGrpcInvoke(serviceType, app.Logger())
		app.appoptions = append(app.appoptions, fx.Invoke(invokeFn))
	}

	return app
}

func makeGrpcInvoke(serviceType reflect.Type, logger *zap.Logger) interface{} {
	// 构造函数类型：func(*grpc.Server, <YourServiceType>)
	fnType := reflect.FuncOf(
		[]reflect.Type{reflect.TypeOf((*grpc.Server)(nil)), serviceType}, // 入参类型
		[]reflect.Type{}, // 返回值类型为空
		false,            // 非变长参数
	)

	// 构造函数实现
	fn := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		server := args[0].Interface().(*grpc.Server)
		svc := args[1].Interface()

		grpcSvc, ok := svc.(GrpcService)
		if !ok {
			panic(fmt.Sprintf("MapGrpcServices: %s does not implement GrpcService", reflect.TypeOf(svc)))
		}

		grpcSvc.Register(server)
		logger.Info("Registered gRPC service", zap.String("type", reflect.TypeOf(svc).String()))

		return nil
	})

	return fn.Interface()
}
