package gin

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

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

// WebApplication 实现 WebApplication 接口
type WebApplication struct {
	*app.Application
	routeRegistrations      []any
	grpcServiceConstructors []any
	container               []fx.Option
	handler                 http.Handler
	router                  web.Router
	server                  *http.Server
	ServerOptions           *web.ServerConfig
	logger                  *zap.Logger
	config                  *viper.Viper
	env                     *web.Environment
}

// NewWebApplication 创建一个 WebApplication 实例
func NewWebApplication(cfg web.InstanceConfig) web.Application {
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

		e.Use(newRecoveryWithZap(cfg.Logger))
	}

	// 6. logger 默认启用（除非明确配置为 false）
	if serverOptions.UseDefaultLogger = !cfg.Config.IsSet("server.use_default_logger") ||
		cfg.Config.GetBool("server.use_default_logger"); serverOptions.UseDefaultLogger {

		e.Use(newZapLogger(cfg.Logger))
	}

	return &WebApplication{
		handler:       e,
		ServerOptions: serverOptions,
		config:        cfg.Config,
		logger:        cfg.Logger,
		container:     cfg.Container,
		env:           env,
		Application:   cfg.Applicaton,
		router:        cfg.Router,
	}
}

// Run 启动 Web 应用程序
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
	webapp.container = append(webapp.container,
		fx.Supply(webapp.handler.(*gin.Engine)),

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
		webapp.container = append(webapp.container,
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
			webapp.container = append(webapp.container, fx.Provide(constructor))
			constructorType := reflect.TypeOf(constructor)
			serviceType := constructorType.Out(0)
			invokeFn := makeGrpcInvoke(serviceType, webapp.logger)
			webapp.container = append(webapp.container, fx.Invoke(invokeFn))
		}
	}

	// 注册 HTTP 路由
	for _, r := range webapp.routeRegistrations {
		webapp.container = append(webapp.container, fx.Invoke(r))
	}

	// 构建并运行 Fx 应用
	webapp.App = fx.New(webapp.container...)

	// 直接使用 Fx 的 Run 来管理生命周期和信号
	webapp.logger.Info("Starting application...")
	webapp.App.Run()
	webapp.logger.Info("Application stopped gracefully")
}

// MapRoutes 注册路由
func (a *WebApplication) MapRoute(routeFuncList ...any) web.Application {

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
func (a *WebApplication) UseSwagger() web.Application {
	a.engine().GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return a
}

// UseCORS 配置跨域
func (a *WebApplication) UseCORS(fn any) web.Application {
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
func (a *WebApplication) UseStaticFiles(urlPath, root string) web.Application {
	a.engine().Static(urlPath, root)
	return a
}

// UseHealthCheck 配置健康检查
func (a *WebApplication) UseHealthCheck() web.Application {
	a.engine().GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	return a
}

func (a *WebApplication) engine() *gin.Engine {
	return a.handler.(*gin.Engine)
}

// MapGrpcServices 注册 gRPC 服务
func (webapp *WebApplication) MapGrpcServices(constructors ...any) web.Application {
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
func (b *WebApplication) Use(middleware ...any) web.Application {
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

// Logger 获取日志实例
func (a *WebApplication) Logger() *zap.Logger {
	return a.logger
}

// Config 获取配置实例
func (a *WebApplication) Config() *viper.Viper {
	return a.config
}

// Environment 获取环境实例
func (a *WebApplication) Env() *web.Environment {
	return a.env
}

// UseRecovery 注册恢复中间件, 用于捕获 panic 并返回 500 错误
func (a *WebApplication) UseRecovery() web.Application {
	a.engine().Use(newRecoveryWithZap(a.logger))
	return a
}

// UseLogger 注册日志中间件, 用于记录请求日志
func (a *WebApplication) UseLogger() web.Application {
	a.engine().Use(newZapLogger(a.logger))
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

func (a *WebApplication) UseRouting() web.Application {
	if a.router == nil {
		panic("RouterOptions is required. Please configure it in WebApplicationOptions.")
	}

	// 注册路由处理器
	a.registerRoutes()

	return a
}

func (a *WebApplication) registerRoutes() {
	// 注册顶级路由
	for _, route := range a.router.Config() {
		if route.Handler == nil {
			continue
		}

		handler := a.CreateRouteInitializer(route.Handler, "", route.Path, route.Method)

		a.routeRegistrations = append(a.routeRegistrations, handler)
	}

	// 注册组路由
	for _, group := range a.router.GroupConfig() {
		for _, route := range group.Routes {
			if route.Handler == nil {
				continue
			}

			handler := a.CreateRouteInitializer(route.Handler, group.Prefix, route.Path, route.Method)

			a.routeRegistrations = append(a.routeRegistrations, handler)
		}
	}
}

func (a *WebApplication) CreateRouteInitializer(handlerFunc any, group, path string, method web.RequestMethod) any {
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
			case web.GET:
				group.GET(path, handler.Interface().(gin.HandlerFunc))
			case web.POST:
				group.POST(path, handler.Interface().(gin.HandlerFunc))
			case web.PUT:
				group.PUT(path, handler.Interface().(gin.HandlerFunc))
			case web.DELETE:
				group.DELETE(path, handler.Interface().(gin.HandlerFunc))
			case web.PATCH:
				group.PATCH(path, handler.Interface().(gin.HandlerFunc))
			}

		} else {

			switch method {
			case web.GET:
				engine.GET(path, handler.Interface().(gin.HandlerFunc))
			case web.POST:
				engine.POST(path, handler.Interface().(gin.HandlerFunc))
			case web.PUT:
				engine.PUT(path, handler.Interface().(gin.HandlerFunc))
			case web.DELETE:
				engine.DELETE(path, handler.Interface().(gin.HandlerFunc))
			case web.PATCH:
				engine.PATCH(path, handler.Interface().(gin.HandlerFunc))
			}

		}
		return nil
	})

	return returnFunc.Interface()
}
