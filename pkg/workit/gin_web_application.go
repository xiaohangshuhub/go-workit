package workit

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	stdstrings "strings"
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

const (
	port        = "8080"
	grpc_port   = "50051"
	environment = "development"
)

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

func newGinWebApplication(options WebApplicationOptions) WebApplication {

	serverOptions := &ServerOptions{
		HttpPort:    port,
		GrpcPort:    grpc_port,
		Environment: environment,
	}

	if err := options.Config.UnmarshalKey("server", &serverOptions); err != nil {
		options.Logger.Error("unmarshal server options failed", zap.Error(err))
	}

	env := &Environment{
		Env:           serverOptions.Environment,
		IsDevelopment: serverOptions.Environment == "development",
	}

	switch stdstrings.ToLower(serverOptions.Environment) {
	case "development":
		gin.SetMode(gin.DebugMode)
	case "testing":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}

	gin := gin.New()
	// 🔥 挂载自己的 zap logger + recovery
	gin.Use(newGinZapLogger(options.Logger))

	gin.Use(recoveryWithZap(options.Logger))

	return &GinWebApplication{
		handler:       gin,
		ServerOptions: serverOptions,
		config:        options.Config,
		logger:        options.Logger,
		container:     options.Container,
		env:           env,
	}
}

func (webapp *GinWebApplication) Run(ctx ...context.Context) error {
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
		return fmt.Errorf("start host failed: %w", err)
	}

	// 等待上下文被取消
	<-appCtx.Done()

	// 优雅关闭服务器
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := webapp.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server failed: %w", err)
	}

	return webapp.app.Stop(shutdownCtx)
}

func (a *GinWebApplication) MapRoutes(registerFunc interface{}) WebApplication {

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

func (b *GinWebApplication) UseMiddleware(constructors ...interface{}) WebApplication {
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
			if !mw.ShouldSkip(c.Request.URL.Path,c.Request.Method) {
				mw.Handle()(c)
			} else {
				c.Next()
			}
		})

		return nil
	})

	return fn.Interface()
}

// 鉴权中间件
func (a *GinWebApplication) UseAuthentication() WebApplication {

	a.UseMiddleware(newGinAuthenticationMiddleware)
	return a
}

// 授权中间件
func (a *GinWebApplication) UseAuthorization() WebApplication {

	a.UseMiddleware(newGinAuthorizationMiddleware)
	return a
}

func (a *GinWebApplication) Logger() *zap.Logger {
	return a.logger
}
func (a *GinWebApplication) Config() *viper.Viper {
	return a.config
}
func (a *GinWebApplication) Env() *Environment {
	return a.env
}
