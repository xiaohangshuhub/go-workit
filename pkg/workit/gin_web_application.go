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

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// GinWebApplication 实现 WebApplication 接口
type GinWebApplication struct {
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

// NewGinWebApplication 创建一个 GinWebApplication 实例
func newGinWebApplication(options WebApplicationOptions) WebApplication {
	serverOptions := &ServerOptions{}

	// 1. http_port 默认 8080
	httpPort := options.Config.GetInt("server.http_port")
	if httpPort == 0 {
		httpPort = 8080
	}
	if httpPort <= 0 || httpPort > 65535 {
		panic(fmt.Sprintf("invalid http_port: %d", httpPort))
	}
	serverOptions.HttpPort = strconv.Itoa(httpPort)

	// 2. grpc_port 默认 50051
	grpcPort := options.Config.GetInt("server.grpc_port")
	if grpcPort == 0 {
		grpcPort = 50051
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

	// 4. 设置 Gin 模式
	switch serverOptions.Environment {
	case development:
		gin.SetMode(gin.DebugMode)
	case testing:
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}

	env := &Environment{
		Env:           serverOptions.Environment,
		IsDevelopment: serverOptions.Environment == development,
	}

	e := gin.New()

	// 5. recover 默认启用（除非明确配置为 false）
	if serverOptions.UseDefaultRecover = !options.Config.IsSet("server.use_default_recover") ||
		options.Config.GetBool("server.use_default_recover"); serverOptions.UseDefaultRecover {

		e.Use(newGinRecoveryWithZap(options.Logger))
	}

	// 6. logger 默认启用（除非明确配置为 false）
	if serverOptions.UseDefaultLogger = !options.Config.IsSet("server.use_default_logger") ||
		options.Config.GetBool("server.use_default_logger"); serverOptions.UseDefaultLogger {

		e.Use(newGinZapLogger(options.Logger))
	}

	return &GinWebApplication{
		handler:       e,
		ServerOptions: serverOptions,
		config:        options.Config,
		logger:        options.Logger,
		container:     options.Container,
		env:           env,
	}
}

// Run 启动 Web 应用程序
func (webapp *GinWebApplication) Run() {

	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 捕获系统信号，优雅关闭
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
			panic(fmt.Errorf("HTTP server ListenAndServe error: %w", err))
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

	for _, r := range webapp.routeRegistrations {
		webapp.container = append(webapp.container, fx.Invoke(r))
	}

	webapp.container = append(webapp.container,
		fx.Supply(webapp.handler.(*gin.Engine)),
	)

	webapp.app = fx.New(webapp.container...)

	// 启动应用程序
	if err := webapp.app.Start(appCtx); err != nil {
		panic(fmt.Errorf("start host failed: %w", err))
	}

	// 等待上下文被取消
	<-appCtx.Done()

	// 优雅关闭服务器
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := webapp.server.Shutdown(shutdownCtx); err != nil {
		panic(fmt.Errorf("shutdown server failed: %w", err))
	}

	if err := webapp.app.Stop(shutdownCtx); err != nil {
		panic(fmt.Errorf("stop host failed: %w", err))
	}
}

// MapRoutes 注册路由
func (a *GinWebApplication) MapRouter(registerFunc interface{}) WebApplication {

	t := reflect.TypeOf(registerFunc)

	if t.Kind() != reflect.Func {
		panic("registerFunc must be a function")
	}

	ginType := reflect.TypeOf(&gin.Engine{})
	hasGin := false

	for i := 0; i < t.NumIn(); i++ {
		if t.In(i) == ginType {
			hasGin = true
			break
		}
	}

	if !hasGin {
		panic("registerFunc must have at least one parameter of type *gin.Engine")
	}

	a.routeRegistrations = append(a.routeRegistrations, registerFunc)
	return a
}

// UseSwagger 配置Swagger
func (a *GinWebApplication) UseSwagger() WebApplication {
	a.engine().GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return a
}

// UseCORS 配置跨域
func (a *GinWebApplication) UseCORS(fn interface{}) WebApplication {
	exec, ok := fn.(func(*cors.Config))
	if !ok {
		panic("UseCORS: argument must be func(*cors.Config)")
	}

	cfg := cors.DefaultConfig()
	exec(&cfg)

	a.engine().Use(cors.New(cfg))
	return a
}

// UseStaticFiles 配置静态文件
func (a *GinWebApplication) UseStaticFiles(urlPath, root string) WebApplication {
	a.engine().Static(urlPath, root)
	return a
}

// UseHealthCheck 配置健康检查
func (a *GinWebApplication) UseHealthCheck() WebApplication {
	a.engine().GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	return a
}

func (a *GinWebApplication) engine() *gin.Engine {
	return a.handler.(*gin.Engine)
}

// MapGrpcServices 注册 gRPC 服务
func (webapp *GinWebApplication) MapGrpcServices(constructors ...interface{}) WebApplication {
	for _, constructor := range constructors {
		webapp.grpcServiceConstructors = append(webapp.grpcServiceConstructors, constructor)
		webapp.container = append(webapp.container, fx.Provide(constructor))

		// 推断构造函数的返回类型
		constructorType := reflect.TypeOf(constructor)
		if constructorType.Kind() != reflect.Func || constructorType.NumOut() == 0 {
			panic("MapGrpcServices: constructor must be a function with at least one return value")
		}

		serviceType := constructorType.Out(0)

		// 对每个具体服务构造出一个 fx.Invoke
		invokeFn := makeGrpcInvoke(serviceType, webapp.logger)
		webapp.container = append(webapp.container, fx.Invoke(invokeFn))
	}

	return webapp
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

// UseMiddleware 注册中间件
func (b *GinWebApplication) Use(constructors ...interface{}) WebApplication {
	for _, constructor := range constructors {
		b.container = append(b.container, fx.Provide(constructor))

		constructorType := reflect.TypeOf(constructor)
		if constructorType.Kind() != reflect.Func || constructorType.NumOut() == 0 {
			panic("UseMiddlewareDI: constructor must be a function that returns Middleware")
		}

		middlewareType := constructorType.Out(0)

		// 生成 fx.Invoke(fn(mwType, *gin.Engine))
		b.container = append(b.container, fx.Invoke(makeMiddlewareInvoke(middlewareType)))
	}
	return b
}

func makeMiddlewareInvoke(middlewareType reflect.Type) interface{} {
	fnType := reflect.FuncOf(
		[]reflect.Type{middlewareType, reflect.TypeOf((*gin.Engine)(nil))},
		[]reflect.Type{},
		false,
	)

	fn := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		mwVal := args[0]
		engine := args[1].Interface().(*gin.Engine)

		mw, ok := mwVal.Interface().(GinMiddleware)
		if !ok {
			panic(fmt.Sprintf("type %v does not implement Middleware", mwVal.Type()))
		}

		engine.Use(func(c *gin.Context) {
			if !mw.ShouldSkip(c.Request.URL.Path, c.Request.Method) {
				mw.Handle()(c)
			} else {
				c.Next()
			}
		})

		return nil
	})

	return fn.Interface()
}

// UseAuthentication 鉴权中间件
func (a *GinWebApplication) UseAuthentication() WebApplication {

	a.Use(newGinAuthenticationMiddleware)
	return a
}

// UseAuthorization 授权中间件
func (a *GinWebApplication) UseAuthorization() WebApplication {

	a.Use(newGinAuthorizationMiddleware)
	return a
}

// Logger 获取日志实例
func (a *GinWebApplication) Logger() *zap.Logger {
	return a.logger
}

// Config 获取配置实例
func (a *GinWebApplication) Config() *viper.Viper {
	return a.config
}

// Environment 获取环境实例
func (a *GinWebApplication) Env() *Environment {
	return a.env
}

// UseRecovery 注册恢复中间件, 用于捕获 panic 并返回 500 错误
func (a *GinWebApplication) UseRecovery() WebApplication {
	a.engine().Use(newGinRecoveryWithZap(a.logger))
	return a
}

// UseLogger 注册日志中间件, 用于记录请求日志
func (a *GinWebApplication) UseLogger() WebApplication {
	a.engine().Use(newGinZapLogger(a.logger))
	return a
}
