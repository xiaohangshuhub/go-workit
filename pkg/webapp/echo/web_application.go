package echo

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/xiaohangshuhub/go-workit/pkg/app"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/router"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/rpc"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// WebApplication 实现 WebApplication 接口
type WebApplication struct {
	*app.Application
	handler                 http.Handler
	server                  *http.Server
	routeRegistrations      []any
	grpcServiceConstructors []any
	ServerOptions           *web.ServerConfig
	env                     *web.Environment
	router                  web.Router
}

// NewWebApplication 创建一个新的 WebApplication
func NewWebApplication(app *app.Application, router *router.Router) web.Application {

	serverOptions := &web.ServerConfig{}

	// 1. http_port 默认 8080
	httpPort := app.Config().GetInt("server.http_port")
	if httpPort == 0 {
		httpPort = web.Port
	}
	if httpPort <= 0 || httpPort > 65535 {
		panic(fmt.Sprintf("invalid http_port: %d", httpPort))
	}
	serverOptions.HttpPort = strconv.Itoa(httpPort)

	// 2. grpc_port 默认 50051
	grpcPort := app.Config().GetInt("server.grpc_port")
	if grpcPort == 0 {
		grpcPort = web.GRPCPort
	}
	if grpcPort <= 0 || grpcPort > 65535 {
		panic(fmt.Sprintf("invalid grpc_port: %d", grpcPort))
	}
	serverOptions.GrpcPort = strconv.Itoa(grpcPort)

	// 3. environment 默认 prod
	environment := strings.ToLower(app.Config().GetString("server.environment"))
	if environment == "" {
		environment = "prod"
	}
	switch environment {
	case web.Development, web.Testing, web.Production:
		serverOptions.Environment = environment
	default:
		panic("invalid environment: " + environment)
	}

	env := &web.Environment{
		Env:           serverOptions.Environment,
		IsDevelopment: serverOptions.Environment == web.Development,
	}

	// 4. 初始化 Echo
	e := echo.New()

	// 开发 / 测试 / 生产模式切换
	switch serverOptions.Environment {
	case web.Development:
		env.IsDevelopment = true
		e.Debug = true
		e.HideBanner = false
		e.HidePort = false
		app.Logger().Info("Running in Debug mode")
	case web.Testing:
		e.Debug = true
		e.HideBanner = true
		e.HidePort = true
		app.Logger().Info("Running in Test mode")
	default: // prod
		e.Debug = false
		e.HideBanner = true
		e.HidePort = true
		app.Logger().Info("Running in Release mode")
	}

	// 5. recover 默认启用（除非明确配置为 false）
	if serverOptions.UseDefaultRecover = !app.Config().IsSet("server.use_default_recover") ||
		app.Config().GetBool("server.use_default_recover"); serverOptions.UseDefaultRecover {
		e.Use(newRecoveryWithZap(app.Logger()))
	}

	// 6. logger 默认启用（除非明确配置为 false）
	if serverOptions.UseDefaultLogger = !app.Config().IsSet("server.use_default_logger") ||
		app.Config().GetBool("server.use_default_logger"); serverOptions.UseDefaultLogger {
		e.Use(newZapLogger(app.Logger(), env.IsDevelopment))
	}

	return &WebApplication{
		handler:       e,
		ServerOptions: serverOptions,
		env:           env,
		Application:   app,
		router:        router,
	}
}

// Run 启动 Web 应用
func (webapp *WebApplication) Run(params ...string) {
	// HTTP server
	webapp.server = &http.Server{
		Addr:         ":" + webapp.ServerOptions.HttpPort,
		Handler:      webapp.handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Fx 容器配置
	webapp.AppendContainer(
		fx.Supply(webapp.handler.(*echo.Echo)),

		// HTTP 生命周期管理
		fx.Invoke(func(lc fx.Lifecycle, shutdowner fx.Shutdowner, logger *zap.Logger) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						logger.Info("HTTP server starting",
							zap.String("port", webapp.ServerOptions.HttpPort))
						if err := webapp.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
							logger.Error("HTTP server error", zap.Error(err))
							_ = shutdowner.Shutdown()
						}
					}()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					logger.Info("Shutting down HTTP server")
					return webapp.server.Shutdown(ctx)
				},
			})
		}),
	)

	// gRPC server 生命周期管理（如果启用）
	if len(webapp.grpcServiceConstructors) > 0 {
		webapp.AppendContainer(
			fx.Provide(func() *grpc.Server {
				return grpc.NewServer()
			}),
			fx.Invoke(func(lc fx.Lifecycle, shutdowner fx.Shutdowner, logger *zap.Logger, grpcSrv *grpc.Server) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						go func() {
							addr := ":" + webapp.ServerOptions.GrpcPort
							lis, err := net.Listen("tcp", addr)
							if err != nil {
								logger.Error("Failed to listen on GRPC port", zap.Error(err))
								_ = shutdowner.Shutdown()
								return
							}
							logger.Info("GRPC server starting", zap.String("port", webapp.ServerOptions.GrpcPort))
							if err := grpcSrv.Serve(lis); err != nil && err != grpc.ErrServerStopped {
								logger.Error("GRPC server error", zap.Error(err))
								_ = shutdowner.Shutdown()
							}
						}()
						return nil
					},
					OnStop: func(ctx context.Context) error {
						logger.Info("Stopping GRPC server")
						stopped := make(chan struct{})
						go func() {
							grpcSrv.GracefulStop()
							close(stopped)
						}()
						select {
						case <-stopped:
							return nil
						case <-ctx.Done():
							grpcSrv.Stop() // 强制关闭
							return ctx.Err()
						}
					},
				})
			}),
		)

		// 注册 gRPC 服务
		for _, constructor := range webapp.grpcServiceConstructors {
			webapp.AppendContainer(fx.Provide(constructor))
			constructorType := reflect.TypeOf(constructor)
			serviceType := constructorType.Out(0)
			invokeFn := makeGrpcInvoke(serviceType, webapp.Logger())
			webapp.AppendContainer(fx.Invoke(invokeFn))
		}
	}

	// 注册 HTTP 路由
	for _, r := range webapp.routeRegistrations {
		webapp.AppendContainer(fx.Invoke(r))
	}

	// 构建并运行 Fx 应用
	fxapp := webapp.FxApp(fx.New(webapp.Container()...))

	// 直接使用 Fx 的 Run 来管理生命周期和信号
	webapp.Logger().Info("Starting application...")
	fxapp.Run()
	webapp.Logger().Info("Application stopped gracefully")
}

// UseStaticFiles 配置静态文件
func (a *WebApplication) UseStaticFiles(urlPath, root string) web.Application {
	a.engine().Static(urlPath, root)
	return a
}

// UseHealthCheck 健康检查
func (a *WebApplication) UseHealthCheck() web.Application {
	a.engine().GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	return a
}

// UseSwagger swagger支持
func (a *WebApplication) UseSwagger() web.Application {
	a.engine().GET("/swagger/*", echoSwagger.WrapHandler)
	return a
}

// UseCORS CORS 支持
func (a *WebApplication) UseCORS(fn any) web.Application {
	exec, ok := fn.(func(*middleware.CORSConfig))
	if !ok {
		panic("UseCORS: argument must be func(*middleware.CORSConfig)")
	}

	cfg := middleware.CORSConfig{}
	exec(&cfg)

	a.engine().Use(middleware.CORSWithConfig(cfg))
	return a
}

// MapRoutes 路由注册
func (a *WebApplication) MapRoute(routeFuncList ...any) web.Application {
	for _, routeFunc := range routeFuncList {
		t := reflect.TypeOf(routeFunc)

		if t.Kind() != reflect.Func {
			panic("registerFunc must be a function")
		}

		echoType := reflect.TypeOf(&echo.Echo{})
		hasEcho := false

		for i := 0; i < t.NumIn(); i++ {
			if t.In(i) == echoType {
				hasEcho = true
				break
			}
		}

		if !hasEcho {
			panic("registerFunc must have at least one parameter of type *echo.Echo")
		}

		a.routeRegistrations = append(a.routeRegistrations, routeFunc)
	}

	return a
}

// engine 返回 echo.Echo 对象
func (a *WebApplication) engine() *echo.Echo {
	return a.handler.(*echo.Echo)
}

// MapGrpcServices 注册 gRPC 服务
func (app *WebApplication) MapGrpcServices(constructors ...any) web.Application {
	for _, constructor := range constructors {
		app.grpcServiceConstructors = append(app.grpcServiceConstructors, constructor)
		app.AppendContainer(fx.Provide(constructor))

		constructorType := reflect.TypeOf(constructor)
		if constructorType.Kind() != reflect.Func || constructorType.NumOut() == 0 {
			panic("MapGrpcServices: constructor must be a function with at least one return value")
		}

		serviceType := constructorType.Out(0)
		invokeFn := makeGrpcInvoke(serviceType, app.Logger())
		app.AppendContainer(fx.Invoke(invokeFn))
	}

	return app
}

func makeGrpcInvoke(serviceType reflect.Type, logger *zap.Logger) any {
	fnType := reflect.FuncOf(
		[]reflect.Type{reflect.TypeOf((*grpc.Server)(nil)), serviceType},
		[]reflect.Type{},
		false,
	)

	fn := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		server := args[0].Interface().(*grpc.Server)
		svc := args[1].Interface()

		grpcSvc, ok := svc.(rpc.GrpcService)
		if !ok {
			panic(fmt.Sprintf("MapGrpcServices: %s does not implement GrpcService", reflect.TypeOf(svc)))
		}

		grpcSvc.Register(server)
		logger.Info("Registered gRPC service", zap.String("type", reflect.TypeOf(svc).String()))

		return nil
	})

	return fn.Interface()
}

// UseMiddleware 注册中间件
func (b *WebApplication) Use(middleware ...any) web.Application {
	for _, constructor := range middleware {
		b.AppendContainer(fx.Provide(constructor))

		constructorType := reflect.TypeOf(constructor)
		if constructorType.Kind() != reflect.Func || constructorType.NumOut() == 0 {
			panic("UseMiddlewareDI: constructor must be a function that returns Middleware")
		}

		middlewareType := constructorType.Out(0)
		b.AppendContainer(fx.Invoke(echoMakeMiddlewareInvoke(middlewareType)))
	}
	return b
}

func echoMakeMiddlewareInvoke(middlewareType reflect.Type) any {
	fnType := reflect.FuncOf(
		[]reflect.Type{middlewareType, reflect.TypeOf((*echo.Echo)(nil))},
		[]reflect.Type{},
		false,
	)

	fn := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		mwVal := args[0]
		engine := args[1].Interface().(*echo.Echo)

		mw, ok := mwVal.Interface().(Middleware)
		if !ok {
			panic(fmt.Sprintf("type %v does not implement Middleware", mwVal.Type()))
		}

		engine.Use(mw.Handle())
		return nil
	})

	return fn.Interface()
}

// UseAuthentication 鉴权中间件
func (a *WebApplication) UseAuthentication() web.Application {
	a.Use(newAuthenticate)
	return a
}

// UseAuthorization 授权中间件
func (a *WebApplication) UseAuthorization() web.Application {
	a.Use(newAuthorize)
	return a
}

// Environment 获取环境对象
func (a *WebApplication) Env() *web.Environment {
	return a.env
}

// UseRecovery 注册恢复中间件
func (a *WebApplication) UseRecovery() web.Application {
	a.engine().Use(newRecoveryWithZap(a.Logger()))
	return a
}

// UseLogger 注册日志中间件
func (a *WebApplication) UseLogger() web.Application {
	a.engine().Use(newZapLogger(a.Logger(), a.env.IsDevelopment))
	return a
}

// UseLocalization 配置国际化功能
func (a *WebApplication) UseLocalization() web.Application {
	a.Use(newLocalization)
	return a
}

// UseRateLimit 配置限流功能
func (a *WebApplication) UseRateLimiter() web.Application {
	a.Use(newRateLimiter)
	return a
}

func (a *WebApplication) UseRequestDecompression() web.Application {
	a.Use(newRateLimiter)
	return a
}

// UseRouting 配置路由
func (a *WebApplication) UseRouting() web.Application {
	if a.router == nil {
		panic("RouterOptions is required. Please configure it in WebApplicationOptions.")
	}

	for _, route := range a.router.Config() {
		if route.Handler == nil {
			continue
		}
		handler := a.CreateRouteInitializer(route.Handler, "", route.Path, route.Method)
		a.routeRegistrations = append(a.routeRegistrations, handler)
	}

	for _, group := range a.router.GroupConfig() {
		for _, route := range group.Routes {
			if route.Handler == nil {
				continue
			}
			handler := a.CreateRouteInitializer(route.Handler, group.Prefix, route.Path, route.Method)
			a.routeRegistrations = append(a.routeRegistrations, handler)
		}
	}

	return a
}

func (a *WebApplication) CreateRouteInitializer(handlerFunc any, group, path string, method web.RequestMethod) any {
	handlerType := reflect.TypeOf(handlerFunc)
	if handlerType.Kind() != reflect.Func {
		panic("handlerFunc必须是函数")
	}

	paramTypes := make([]reflect.Type, 0, handlerType.NumIn()+1)
	paramTypes = append(paramTypes, reflect.TypeOf(&echo.Echo{}))
	for i := 0; i < handlerType.NumIn(); i++ {
		paramTypes = append(paramTypes, handlerType.In(i))
	}

	returnFuncType := reflect.FuncOf(paramTypes, []reflect.Type{}, false)
	returnFunc := reflect.MakeFunc(returnFuncType, func(args []reflect.Value) []reflect.Value {
		engine := args[0].Interface().(*echo.Echo)
		handlerArgs := args[1:]

		handler := reflect.ValueOf(handlerFunc).Call(handlerArgs)[0]

		if group != "" {
			g := engine.Group(group)
			switch method {
			case web.GET:
				g.GET(path, handler.Interface().(echo.HandlerFunc))
			case web.POST:
				g.POST(path, handler.Interface().(echo.HandlerFunc))
			case web.PUT:
				g.PUT(path, handler.Interface().(echo.HandlerFunc))
			case web.DELETE:
				g.DELETE(path, handler.Interface().(echo.HandlerFunc))
			case web.PATCH:
				g.PATCH(path, handler.Interface().(echo.HandlerFunc))
			}
		} else {
			switch method {
			case web.GET:
				engine.GET(path, handler.Interface().(echo.HandlerFunc))
			case web.POST:
				engine.POST(path, handler.Interface().(echo.HandlerFunc))
			case web.PUT:
				engine.PUT(path, handler.Interface().(echo.HandlerFunc))
			case web.DELETE:
				engine.DELETE(path, handler.Interface().(echo.HandlerFunc))
			case web.PATCH:
				engine.PATCH(path, handler.Interface().(echo.HandlerFunc))
			}
		}
		return nil
	})

	return returnFunc.Interface()
}
