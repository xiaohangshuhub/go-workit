package workit

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
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// EchoWebApplication 实现 WebApplication 接口
type EchoWebApplication struct {
	handler                 http.Handler
	server                  *http.Server
	routeRegistrations      []interface{}
	grpcServiceConstructors []interface{}
	ServerOptions           *ServerOptions
	logger                  *zap.Logger
	config                  *viper.Viper
	container               []fx.Option
	app                     *fx.App
	env                     *Environment
}

// NewEchoWebApplication 创建一个新的 EchoWebApplication
func NewEchoWebApplication(options WebApplicationOptions) WebApplication {
	serverOptions := &ServerOptions{}

	// 1. http_port 默认 8080
	httpPort := options.Config.GetInt("server.http_port")
	if httpPort == 0 {
		httpPort = port
	}
	if httpPort <= 0 || httpPort > 65535 {
		panic(fmt.Sprintf("invalid http_port: %d", httpPort))
	}
	serverOptions.HttpPort = strconv.Itoa(httpPort)

	// 2. grpc_port 默认 50051
	grpcPort := options.Config.GetInt("server.grpc_port")
	if grpcPort == 0 {
		grpcPort = g_port
	}
	if grpcPort <= 0 || grpcPort > 65535 {
		panic(fmt.Sprintf("invalid grpc_port: %d", grpcPort))
	}
	serverOptions.GrpcPort = strconv.Itoa(grpcPort)

	// 3. environment 默认 prod
	environment := strings.ToLower(options.Config.GetString("server.environment"))
	if environment == "" {
		environment = "prod"
	}
	switch environment {
	case development, testing, production:
		serverOptions.Environment = environment
	default:
		panic("invalid environment: " + environment)
	}

	env := &Environment{
		Env:           serverOptions.Environment,
		IsDevelopment: serverOptions.Environment == development,
	}

	// 4. 初始化 Echo
	e := echo.New()

	// 开发 / 测试 / 生产模式切换
	switch serverOptions.Environment {
	case development:
		env.IsDevelopment = true
		e.Debug = true
		e.HideBanner = false
		e.HidePort = false
		options.Logger.Info("Running in Debug mode")
	case testing:
		e.Debug = true
		e.HideBanner = true
		e.HidePort = true
		options.Logger.Info("Running in Test mode")
	default: // prod
		e.Debug = false
		e.HideBanner = true
		e.HidePort = true
		options.Logger.Info("Running in Release mode")
	}

	// 5. recover 默认启用（除非明确配置为 false）
	if serverOptions.UseDefaultRecover = !options.Config.IsSet("server.use_default_recover") ||
		options.Config.GetBool("server.use_default_recover"); serverOptions.UseDefaultRecover {

		e.Use(newEchoRecoveryWithZap(options.Logger))
	}

	// 6. logger 默认启用（除非明确配置为 false）
	if serverOptions.UseDefaultLogger = !options.Config.IsSet("server.use_default_logger") ||
		options.Config.GetBool("server.use_default_logger"); serverOptions.UseDefaultLogger {

		e.Use(newEchoZapLogger(options.Logger, env.IsDevelopment))
	}

	return &EchoWebApplication{
		handler:       e,
		ServerOptions: serverOptions,
		config:        options.Config,
		logger:        options.Logger,
		container:     options.Container,
		env:           env,
	}
}

func (webapp *EchoWebApplication) Run() {

	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		fmt.Println("Received shutdown signal")
		cancel()
	}()

	webapp.server = &http.Server{
		Addr:         ":" + webapp.ServerOptions.HttpPort,
		Handler:      webapp.handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动 HTTP 服务器
	go func() {
		webapp.logger.Info("HTTP server starting...", zap.String("port", webapp.ServerOptions.HttpPort))

		if err := webapp.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			webapp.logger.Error("HTTP server ListenAndServe error", zap.Error(err))
			panic(err)
		}

	}()

	// 启动 gRPC 服务器
	if len(webapp.grpcServiceConstructors) > 0 {

		webapp.container = append(webapp.container,
			fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger) *grpc.Server {
				return NewGrpcServer(lc, logger, *webapp.ServerOptions)
			}),
		)
	}

	// 注册路由
	for _, r := range webapp.routeRegistrations {
		webapp.container = append(webapp.container, fx.Invoke(r))
	}

	webapp.container = append(webapp.container,
		fx.Supply(webapp.handler.(*echo.Echo)), // echo.Echo 实现 http.Handler
	)

	webapp.app = fx.New(webapp.container...)

	if err := webapp.app.Start(appCtx); err != nil {
		panic(fmt.Errorf("start host failed: %w", err))
	}

	<-appCtx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := webapp.server.Shutdown(shutdownCtx); err != nil {
		panic(fmt.Errorf("shutdown server failed: %w", err))
	}

	if err := webapp.app.Stop(shutdownCtx); err != nil {
		panic(fmt.Errorf("stop host failed: %w", err))
	}
}

// UseStaticFiles 配置静态文件
func (a *EchoWebApplication) UseStaticFiles(urlPath, root string) WebApplication {
	a.engine().Static(urlPath, root)
	return a
}

// UseHealthCheck 健康检查
func (a *EchoWebApplication) UseHealthCheck() WebApplication {
	a.engine().GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	return a
}

// UseSwagger swagger支持
func (a *EchoWebApplication) UseSwagger() WebApplication {
	a.engine().GET("/swagger/*", echoSwagger.WrapHandler)
	return a
}

// UseCORS CORS 支持
func (a *EchoWebApplication) UseCORS(fn interface{}) WebApplication {
	// 断言传入参数为 func(*middleware.CORSConfig)
	exec, ok := fn.(func(*middleware.CORSConfig))
	if !ok {
		panic("UseCORS: argument must be func(*middleware.CORSConfig)")
	}

	// 取默认配置
	cfg := middleware.CORSConfig{}

	// 调用传入的函数修改配置
	exec(&cfg)

	// 注册中间件
	a.engine().Use(middleware.CORSWithConfig(cfg))

	return a
}

// MapRoutes 路由注册
func (a *EchoWebApplication) MapRouter(registerFunc interface{}) WebApplication {
	t := reflect.TypeOf(registerFunc)

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

	a.routeRegistrations = append(a.routeRegistrations, registerFunc)

	return a
}

// engine 返回 echo.Echo 对象
func (a *EchoWebApplication) engine() *echo.Echo {
	return a.handler.(*echo.Echo)
}

// MapGrpcServices 注册 gRPC 服务
func (app *EchoWebApplication) MapGrpcServices(constructors ...interface{}) WebApplication {
	for _, constructor := range constructors {
		app.grpcServiceConstructors = append(app.grpcServiceConstructors, constructor)
		app.container = append(app.container, fx.Provide(constructor))

		// 推断构造函数的返回类型
		constructorType := reflect.TypeOf(constructor)
		if constructorType.Kind() != reflect.Func || constructorType.NumOut() == 0 {
			panic("MapGrpcServices: constructor must be a function with at least one return value")
		}

		serviceType := constructorType.Out(0)

		// 对每个具体服务构造出一个 fx.Invoke
		invokeFn := echoMakeGrpcInvoke(serviceType, app.logger)
		app.container = append(app.container, fx.Invoke(invokeFn))
	}

	return app
}

// echoMakeGrpcInvoke 构造一个 fx.Invoke 用于注册 gRPC 服务
func echoMakeGrpcInvoke(serviceType reflect.Type, logger *zap.Logger) interface{} {
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

// UseMiddleware 注册中间件
func (b *EchoWebApplication) Use(constructors ...interface{}) WebApplication {
	for _, constructor := range constructors {
		b.container = append(b.container, fx.Provide(constructor))

		constructorType := reflect.TypeOf(constructor)
		if constructorType.Kind() != reflect.Func || constructorType.NumOut() == 0 {
			panic("UseMiddlewareDI: constructor must be a function that returns Middleware")
		}

		middlewareType := constructorType.Out(0)

		// 生成 fx.Invoke(fn(mwType, *gin.Engine))
		b.container = append(b.container, fx.Invoke(echoMakeMiddlewareInvoke(middlewareType)))
	}
	return b
}

// echoMakeMiddlewareInvoke 构造一个 fx.Invoke 用于注册中间件
func echoMakeMiddlewareInvoke(middlewareType reflect.Type) interface{} {
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

		engine.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				if mw.ShouldSkip(c.Request().URL.Path, c.Request().Method) {
					return next(c)
				}
				return mw.Handle()(next)(c) // 注意这里传 next
			}
		})

		return nil
	})

	return fn.Interface()
}

// UseAuthentication 鉴权中间件
func (a *EchoWebApplication) UseAuthentication() WebApplication {

	a.Use(newEchoAuthenticationMiddleware)
	return a
}

// UseAuthorization 授权中间件
func (a *EchoWebApplication) UseAuthorization() WebApplication {

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
func (a *EchoWebApplication) Env() *Environment {
	return a.env
}

// UseRecovery 注册恢复中间件, 用于捕获 panic 并返回 500 错误
func (a *EchoWebApplication) UseRecovery() WebApplication {
	a.engine().Use(newEchoRecoveryWithZap(a.logger))
	return a
}

// UseLogger 注册日志中间件, 用于记录请求日志
func (a *EchoWebApplication) UseLogger() WebApplication {
	a.engine().Use(newEchoZapLogger(a.logger, a.env.IsDevelopment))
	return a
}
