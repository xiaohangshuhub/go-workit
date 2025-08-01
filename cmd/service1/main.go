package main

import (
	"fmt"

	_ "github.com/xiaohangshuhub/go-workit/api/service1/docs" // swagger 一定要有这行,指向你的文档地址
	"github.com/xiaohangshuhub/go-workit/internal/service1/grpcapi/hello"
	"github.com/xiaohangshuhub/go-workit/internal/service1/webapi"
	"github.com/xiaohangshuhub/go-workit/pkg/workit"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// 结构体
type HelloService struct {
	log *zap.Logger
}

// 构建服务
func NewHelloService(log *zap.Logger) *HelloService {
	return &HelloService{log: log}
}

func main() {

	// web应用构建器
	builder := workit.NewWebAppBuilder()

	// 配置构建器(注册即生效)
	builder.AddConfig(func(build workit.ConfigBuilder) {
		build.AddYamlFile("./config.yaml")
	})
	// 配置构建器读取数据
	var port = builder.Config().Get("server.port")
	fmt.Println("server port:", port)

	// 服务注册
	builder.AddServices(fx.Provide(NewHelloService))

	app, err := builder.Build()

	if err != nil {
		fmt.Printf("Failed to build application: %v\n", err)
		return
	}

	if app.Env.IsDevelopment {
		app.UseSwagger()
	}

	// 配置路由
	app.MapRoutes(webapi.Hello)

	// 配置grpc服务
	app.MapGrpcServices(hello.NewHelloService)

	// 运行应用
	if err := app.Run(); err != nil {
		app.Logger().Error("Error running application", zap.Error(err))
	}
}
