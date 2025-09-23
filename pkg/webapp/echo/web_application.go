package echo

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/xiaohangshuhub/go-workit/pkg/app"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/rpc"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// EchoWebApplication 实现 WebApplication 接口
type EchoWebApplication struct {
	*app.Application
	handler                 http.Handler
	server                  *http.Server
	routeRegistrations      []any
	grpcServiceConstructors []any
	ServerOptions           *web.ServerConfig
	logger                  *zap.Logger
	config                  *viper.Viper
	container               []fx.Option
	env                     *web.Environment
	routerConfig            web.RouterConfig
}

// NewEchoWebApplication 创建一个新的 EchoWebApplication
func NewEchoWebApplication(cfg web.InstanceConfig) web.Application {

	serverOptions := &web.ServerConfig{}

	// 1. http_port 默认 8080
	httpPort := cfg.Config.GetInt("server.http_port")
	if httpPort == 0 {
		httpPort = web.Port
	}
	if httpPort <= 0 || httpPort > 65535 {
		panic(fmt.Sprintf("invalid http_port: %d", httpPort))
	}
	serverOptions.HttpPort = strconv.Itoa(httpPort)

	// 2. grpc_port 默认 50051
	grpcPort := cfg.Config.GetInt("server.grpc_port")
	if grpcPort == 0 {
		grpcPort = web.GRPCPort
	}
	if grpcPort <= 0 || grpcPort > 65535 {
		panic(fmt.Sprintf("invalid grpc_port: %d", grpcPort))
	}
	serverOptions.GrpcPort = strconv.Itoa(grpcPort)

	// 3. environment 默认 prod
	environment := strings.ToLower(cfg.Config.GetString("server.environment"))
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
		cfg.Logger.Info("Running in Debug mode")
	case web.Testing:
		e.Debug = true
		e.HideBanner = true
		e.HidePort = true
		cfg.Logger.Info("Running in Test mode")
	default: // prod
		e.Debug = false
		e.HideBanner = true
		e.HidePort = true
		cfg.Logger.Info("Running in Release mode")
	}

	// 5. recover 默认启用（除非明确配置为 false）
	if serverOptions.UseDefaultRecover = !cfg.Config.IsSet("server.use_default_recover") ||
		cfg.Config.GetBool("server.use_default_recover"); serverOptions.UseDefaultRecover {
		e.Use(newEchoRecoveryWithZap(cfg.Logger))
	}

	// 6. logger 默认启用（除非明确配置为 false）
	if serverOptions.UseDefaultLogger = !cfg.Config.IsSet("server.use_default_logger") ||
		cfg.Config.GetBool("server.use_default_logger"); serverOptions.UseDefaultLogger {
		e.Use(newEchoZapLogger(cfg.Logger, env.IsDevelopment))
	}

	return &EchoWebApplication{
		handler:       e,
		ServerOptions: serverOptions,
		config:        cfg.Config,
		logger:        cfg.Logger,
		container:     cfg.Container,
		env:           env,
		Application:   cfg.Applicaton,
		routerConfig:  cfg.RouterConfig,
	}
}

// Run 启动 Web 应用
func (webapp *EchoWebApplication) Run(params ...string) {
	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 捕获信号
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		webapp.logger.Info("Received shutdown signal")
		cancel()
	}()

	// 初始化 HTTP server
	webapp.server = &http.Server{
		Addr:         ":" + webapp.ServerOptions.HttpPort,
		Handler:      webapp.handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 配置 gRPC
	if len(webapp.grpcServiceConstructors) > 0 {
		webapp.container = append(webapp.container,
			fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger) *grpc.Server {
				return rpc.NewGrpcServer(lc, logger, webapp.ServerOptions.GrpcPort)
			}),
		)
	}

	// 注册路由
	for _, r := range webapp.routeRegistrations {
		webapp.container = append(webapp.container, fx.Invoke(r))
	}

	// 交给 Fx 管理
	webapp.container = append(webapp.container,
		fx.Supply(app.NewAppContext(appCtx)),
		fx.Supply(webapp.handler.(*echo.Echo)),
		fx.Invoke(func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						webapp.logger.Info("HTTP server starting...", zap.String("port", webapp.ServerOptions.HttpPort))
						if err := webapp.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
							webapp.logger.Error("HTTP server ListenAndServe error", zap.Error(err))
							cancel()
						}
					}()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					webapp.logger.Info("Shutting down HTTP server")
					return webapp.server.Shutdown(ctx)
				},
			})
		}),
	)

	webapp.App = fx.New(webapp.container...)

	// 启动应用
	webapp.logger.Info("Starting application...")
	if err := webapp.App.Start(appCtx); err != nil {
		webapp.logger.Error("Failed to start application", zap.Error(err))
		return
	}

	webapp.logger.Info("Application started successfully")

	<-appCtx.Done()
	webapp.logger.Info("Application shutdown initiated")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	webapp.logger.Info("Stopping application...")
	if err := webapp.App.Stop(shutdownCtx); err != nil {
		webapp.logger.Error("Failed to stop application gracefully", zap.Error(err))
	} else {
		webapp.logger.Info("Application stopped gracefully")
	}
}

// UseStaticFiles 配置静态文件
func (a *EchoWebApplication) UseStaticFiles(urlPath, root string) web.Application {
	a.engine().Static(urlPath, root)
	return a
}

// UseHealthCheck 健康检查
func (a *EchoWebApplication) UseHealthCheck() web.Application {
	a.engine().GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	return a
}

// UseSwagger swagger支持
func (a *EchoWebApplication) UseSwagger() web.Application {
	a.engine().GET("/swagger/*", echoSwagger.WrapHandler)
	return a
}

// UseCORS CORS 支持
func (a *EchoWebApplication) UseCORS(fn any) web.Application {
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
func (a *EchoWebApplication) MapRouter(routeFuncList ...any) web.Application {
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
func (a *EchoWebApplication) engine() *echo.Echo {
	return a.handler.(*echo.Echo)
}

// MapGrpcServices 注册 gRPC 服务
func (app *EchoWebApplication) MapGrpcServices(constructors ...any) web.Application {
	for _, constructor := range constructors {
		app.grpcServiceConstructors = append(app.grpcServiceConstructors, constructor)
		app.container = append(app.container, fx.Provide(constructor))

		constructorType := reflect.TypeOf(constructor)
		if constructorType.Kind() != reflect.Func || constructorType.NumOut() == 0 {
			panic("MapGrpcServices: constructor must be a function with at least one return value")
		}

		serviceType := constructorType.Out(0)
		invokeFn := echoMakeGrpcInvoke(serviceType, app.logger)
		app.container = append(app.container, fx.Invoke(invokeFn))
	}

	return app
}

func echoMakeGrpcInvoke(serviceType reflect.Type, logger *zap.Logger) any {
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
func (b *EchoWebApplication) Use(middleware ...any) web.Application {
	for _, constructor := range middleware {
		b.container = append(b.container, fx.Provide(constructor))

		constructorType := reflect.TypeOf(constructor)
		if constructorType.Kind() != reflect.Func || constructorType.NumOut() == 0 {
			panic("UseMiddlewareDI: constructor must be a function that returns Middleware")
		}

		middlewareType := constructorType.Out(0)
		b.container = append(b.container, fx.Invoke(echoMakeMiddlewareInvoke(middlewareType)))
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

		mw, ok := mwVal.Interface().(EchoMiddleware)
		if !ok {
			panic(fmt.Sprintf("type %v does not implement Middleware", mwVal.Type()))
		}

		engine.Use(mw.Handle())
		return nil
	})

	return fn.Interface()
}

// UseAuthentication 鉴权中间件
func (a *EchoWebApplication) UseAuthentication() web.Application {
	a.Use(newEchoAuthenticationMiddleware)
	return a
}

// UseAuthorization 授权中间件
func (a *EchoWebApplication) UseAuthorization() web.Application {
	a.Use(newEchoAuthorizationMiddleware)
	return a
}

// Logger 获取日志对象
func (a *EchoWebApplication) Logger() *zap.Logger {
	return a.logger
}

// Config 获取配置对象
func (a *EchoWebApplication) Config() *viper.Viper {
	return a.config
}

// Environment 获取环境对象
func (a *EchoWebApplication) Env() *web.Environment {
	return a.env
}

// UseRecovery 注册恢复中间件
func (a *EchoWebApplication) UseRecovery() web.Application {
	a.engine().Use(newEchoRecoveryWithZap(a.logger))
	return a
}

// UseLogger 注册日志中间件
func (a *EchoWebApplication) UseLogger() web.Application {
	a.engine().Use(newEchoZapLogger(a.logger, a.env.IsDevelopment))
	return a
}

// UseLocalization 配置国际化功能
func (a *EchoWebApplication) UseLocalization() web.Application {
	a.Use(newEchoLocalizationMiddleware)
	return a
}

// UseRateLimit 配置限流功能
func (a *EchoWebApplication) UseRateLimiter() web.Application {
	a.Use(newEchoRateLimitMiddleware)
	return a
}

// UseRouting 配置路由
func (a *EchoWebApplication) UseRouting() web.Application {
	if a.routerConfig == nil {
		panic("RouterOptions is required. Please configure it in WebApplicationOptions.")
	}

	a.registerRoutes()
	return a
}

func (a *EchoWebApplication) registerRoutes() {
	for _, route := range a.routerConfig.RouteConfig() {
		if route.Handler == nil {
			continue
		}
		handler := a.CreateRouteInitializer(route.Handler, "", route.Path, route.Method)
		a.routeRegistrations = append(a.routeRegistrations, handler)
	}

	for _, group := range a.routerConfig.GroupRouteConfig() {
		for _, route := range group.Routes {
			if route.Handler == nil {
				continue
			}
			handler := a.CreateRouteInitializer(route.Handler, group.Prefix, route.Path, route.Method)
			a.routeRegistrations = append(a.routeRegistrations, handler)
		}
	}
}

func (a *EchoWebApplication) CreateRouteInitializer(handlerFunc any, group, path string, method web.RequestMethod) any {
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
