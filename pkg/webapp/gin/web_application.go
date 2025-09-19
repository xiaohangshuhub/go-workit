package gin

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

	"github.com/xiaohangshuhub/go-workit/pkg/webapp/router"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/rpc"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/xiaohangshuhub/go-workit/pkg/app"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// GinWebApplication 实现 WebApplication 接口
type GinWebApplication struct {
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
	routerProvider          web.RouterProvider
}

// NewGinWebApplication 创建一个 GinWebApplication 实例
func NewGinWebApplication(cfg web.InstanceConfig) web.Application {
	serverOptions := &web.ServerConfig{}

	// 1. http_port 默认 8080
	httpPort := cfg.Config.GetInt("server.http_port")
	if httpPort == 0 {
		httpPort = 8080
	}
	if httpPort <= 0 || httpPort > 65535 {
		panic(fmt.Sprintf("invalid http_port: %d", httpPort))
	}
	serverOptions.HttpPort = strconv.Itoa(httpPort)

	// 2. grpc_port 默认 50051
	grpcPort := cfg.Config.GetInt("server.grpc_port")
	if grpcPort == 0 {
		grpcPort = 50051
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

	// 4. 设置 Gin 模式
	switch serverOptions.Environment {
	case web.Development:
		gin.SetMode(gin.DebugMode)
	case web.Testing:
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}

	env := &web.Environment{
		Env:           serverOptions.Environment,
		IsDevelopment: serverOptions.Environment == web.Development,
	}

	e := gin.New()

	// 5. recover 默认启用（除非明确配置为 false）
	if serverOptions.UseDefaultRecover = !cfg.Config.IsSet("server.use_default_recover") ||
		cfg.Config.GetBool("server.use_default_recover"); serverOptions.UseDefaultRecover {

		e.Use(newGinRecoveryWithZap(cfg.Logger))
	}

	// 6. logger 默认启用（除非明确配置为 false）
	if serverOptions.UseDefaultLogger = !cfg.Config.IsSet("server.use_default_logger") ||
		cfg.Config.GetBool("server.use_default_logger"); serverOptions.UseDefaultLogger {

		e.Use(newGinZapLogger(cfg.Logger))
	}

	return &GinWebApplication{
		handler:        e,
		ServerOptions:  serverOptions,
		config:         cfg.Config,
		logger:         cfg.Logger,
		container:      cfg.Container,
		env:            env,
		Application:    cfg.Applicaton,
		routerProvider: cfg.RouterProvider,
	}
}

// Run 启动 Web 应用程序
func (webapp *GinWebApplication) Run(params ...string) {
	// 创建主应用 Context
	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel() // 确保最终会调用 cancel

	// 捕获系统信号，优雅关闭
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		webapp.logger.Info("Received shutdown signal")
		cancel() // 取消主 Context，触发关闭流程
	}()

	// 初始化 HTTP 服务器
	webapp.server = &http.Server{
		Addr:         ":" + webapp.ServerOptions.HttpPort,
		Handler:      webapp.handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 配置grpc服务 Fx 容器
	// if len(webapp.grpcServiceConstructors) > 0 {
	// 	webapp.container = append(webapp.container,
	// 		fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger) *grpc.Server {
	// 			return NewGrpcServer(lc, logger, *webapp.ServerOptions)
	// 		}),
	// 	)
	// }

	for _, r := range webapp.routeRegistrations {
		webapp.container = append(webapp.container, fx.Invoke(r))
	}

	// 添加 HTTP 服务器生命周期管理到 Fx
	webapp.container = append(webapp.container,
		fx.Supply(app.NewAppContext(appCtx)),
		fx.Supply(webapp.handler.(*gin.Engine)),
		// 将 HTTP 服务器纳入 Fx 生命周期管理
		fx.Invoke(func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					// 在 Fx 启动时启动 HTTP 服务器
					go func() {
						webapp.logger.Info("HTTP server starting...",
							zap.String("port", webapp.ServerOptions.HttpPort))
						if err := webapp.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
							webapp.logger.Error("HTTP server ListenAndServe error", zap.Error(err))
							cancel() // 如果服务器启动失败，触发关闭
						}
					}()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					// 在 Fx 停止时关闭 HTTP 服务器
					webapp.logger.Info("Shutting down HTTP server")
					return webapp.server.Shutdown(ctx)
				},
			})
		}),
	)

	webapp.App = fx.New(webapp.container...)

	// 启动应用程序（使用 appCtx）
	webapp.logger.Info("Starting application...")
	if err := webapp.App.Start(appCtx); err != nil {
		webapp.logger.Error("Failed to start application", zap.Error(err))
		return
	}

	webapp.logger.Info("Application started successfully")

	// 等待主 Context 被取消（收到关闭信号）
	<-appCtx.Done()
	webapp.logger.Info("Application shutdown initiated")

	// 创建关闭超时 Context
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// 停止应用程序（使用同一个 appCtx 的派生 Context）
	webapp.logger.Info("Stopping application...")
	if err := webapp.App.Stop(shutdownCtx); err != nil {
		webapp.logger.Error("Failed to stop application gracefully", zap.Error(err))
	} else {
		webapp.logger.Info("Application stopped gracefully")
	}
}

// MapRoutes 注册路由
func (a *GinWebApplication) MapRouter(routeFuncList ...any) web.Application {

	for _, routeFunc := range routeFuncList {

		t := reflect.TypeOf(routeFunc)

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

		a.routeRegistrations = append(a.routeRegistrations, routeFunc)
	}

	return a
}

// UseSwagger 配置Swagger
func (a *GinWebApplication) UseSwagger() web.Application {
	a.engine().GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return a
}

// UseCORS 配置跨域
func (a *GinWebApplication) UseCORS(fn any) web.Application {
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
func (a *GinWebApplication) UseStaticFiles(urlPath, root string) web.Application {
	a.engine().Static(urlPath, root)
	return a
}

// UseHealthCheck 配置健康检查
func (a *GinWebApplication) UseHealthCheck() web.Application {
	a.engine().GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	return a
}

func (a *GinWebApplication) engine() *gin.Engine {
	return a.handler.(*gin.Engine)
}

// MapGrpcServices 注册 gRPC 服务
func (webapp *GinWebApplication) MapGrpcServices(constructors ...any) web.Application {
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

func makeGrpcInvoke(serviceType reflect.Type, logger *zap.Logger) any {
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
func (b *GinWebApplication) Use(middleware ...any) web.Application {
	for _, constructor := range middleware {
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

func makeMiddlewareInvoke(middlewareType reflect.Type) any {
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

		engine.Use(mw.Handle())

		return nil
	})

	return fn.Interface()
}

// UseAuthentication 鉴权中间件
func (a *GinWebApplication) UseAuthentication() web.Application {

	a.Use(newGinAuthenticationMiddleware)
	return a
}

// UseAuthorization 授权中间件
func (a *GinWebApplication) UseAuthorization() web.Application {

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
func (a *GinWebApplication) Env() *web.Environment {
	return a.env
}

// UseRecovery 注册恢复中间件, 用于捕获 panic 并返回 500 错误
func (a *GinWebApplication) UseRecovery() web.Application {
	a.engine().Use(newGinRecoveryWithZap(a.logger))
	return a
}

// UseLogger 注册日志中间件, 用于记录请求日志
func (a *GinWebApplication) UseLogger() web.Application {
	a.engine().Use(newGinZapLogger(a.logger))
	return a
}

// UseLocalization 配置国际化功能
func (a *GinWebApplication) UseLocalization() web.Application {
	a.Use(newGinLocalizationMiddleware)
	return a
}

// UseRateLimit 配置限流功能
func (a *GinWebApplication) UseRateLimiter() web.Application {
	a.Use(newGinRateLimitMiddleware)
	return a
}

func (a *GinWebApplication) UseRouting() web.Application {
	if a.routerProvider == nil {
		panic("RouterOptions is required. Please configure it in WebApplicationOptions.")
	}

	// 注册路由处理器
	a.registerRoutes()

	return a
}

func (a *GinWebApplication) registerRoutes() {
	// 注册顶级路由
	for _, route := range a.routerProvider.RouteConfig() {
		if route.Handler == nil {
			continue
		}

		handler := a.CreateRouteInitializer(route.Handler, "", route.Path, route.Method)

		a.routeRegistrations = append(a.routeRegistrations, handler)
	}

	// 注册组路由
	for _, group := range a.routerProvider.GroupRouteConfig() {
		for _, route := range group.Routes {
			if route.Handler == nil {
				continue
			}

			handler := a.CreateRouteInitializer(route.Handler, group.Prefix, route.Path, route.Method)

			a.routeRegistrations = append(a.routeRegistrations, handler)
		}
	}
}

func (a *GinWebApplication) CreateRouteInitializer(handlerFunc any, group, path string, method router.RequestMethod) any {
	// 获取handler函数的参数类型
	handlerType := reflect.TypeOf(handlerFunc)
	if handlerType.Kind() != reflect.Func {
		panic("handlerFunc必须是函数")
	}

	// 构造返回函数类型: func(*gin.Engine, ...handler参数)
	paramTypes := make([]reflect.Type, 0, handlerType.NumIn()+1)
	paramTypes = append(paramTypes, reflect.TypeOf(&gin.Engine{}))
	for i := 0; i < handlerType.NumIn(); i++ {
		paramTypes = append(paramTypes, handlerType.In(i))
	}

	// 动态创建函数
	returnFuncType := reflect.FuncOf(paramTypes, []reflect.Type{}, false)
	returnFunc := reflect.MakeFunc(returnFuncType, func(args []reflect.Value) []reflect.Value {
		// 提取参数
		engine := args[0].Interface().(*gin.Engine)
		handlerArgs := args[1:]

		// 调用handler工厂函数
		handler := reflect.ValueOf(handlerFunc).Call(handlerArgs)[0]

		if group != "" {

			// 注册路由
			group := engine.Group(group)

			switch method {
			case router.GET:
				group.GET(path, handler.Interface().(gin.HandlerFunc))
			case router.POST:
				group.POST(path, handler.Interface().(gin.HandlerFunc))
			case router.PUT:
				group.PUT(path, handler.Interface().(gin.HandlerFunc))
			case router.DELETE:
				group.DELETE(path, handler.Interface().(gin.HandlerFunc))
			case router.PATCH:
				group.PATCH(path, handler.Interface().(gin.HandlerFunc))
			}

		} else {

			switch method {
			case router.GET:
				engine.GET(path, handler.Interface().(gin.HandlerFunc))
			case router.POST:
				engine.POST(path, handler.Interface().(gin.HandlerFunc))
			case router.PUT:
				engine.PUT(path, handler.Interface().(gin.HandlerFunc))
			case router.DELETE:
				engine.DELETE(path, handler.Interface().(gin.HandlerFunc))
			case router.PATCH:
				engine.PATCH(path, handler.Interface().(gin.HandlerFunc))
			}

		}
		return nil
	})

	return returnFunc.Interface()
}
