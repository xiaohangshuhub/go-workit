// Package main API文档
//
// @title           我的服务 API
// @version         1.0
// @description     这是一个示例 API 文档
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入格式: Bearer {token}
package main

import (
	_ "github.com/xiaohangshuhub/go-workit/api/service1/docs" // swagger 一定要有这行,指向你的文档地址
	"github.com/xiaohangshuhub/go-workit/config"
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

	// 服务注册
	builder.AddServices(fx.Provide(NewHelloService))

	//注册鉴权: 采用 jwt 认证
	builder.AddAuthentication().AddJwtBearer(
		func(options *workit.JwtBearerOptions) {
			options.Authority = "http://localhost:8090"
			options.RequireHttpsMetadata = false
			options.TokenValidationParameters = workit.TokenValidationParameters{
				ValidateIssuer: true,
				ValidIssuer:    "http://localhost:8090",
			}
		})

	// 注册授权策略
	builder.AddAuthorization(config.Authorize...).RequireRole("admin_role_policy", "admin", "super_admin")

	app, err := builder.Build()

	if err != nil {
		app.Logger.Error("Failed to build application: %v\n", zap.Error(err))
		return
	}

	if app.Env.IsDevelopment {
		app.UseSwagger()
	}

	// 配置鉴权
	app.UseAuthentication()

	// 配置授权
	app.UseAuthorization()

	// 配置路由
	app.MapRoutes(webapi.Hello)

	// 配置grpc服务
	app.MapGrpcServices(hello.NewHelloService)

	// 运行应用
	if err := app.Run(); err != nil {
		app.Logger.Error("Error running application", zap.Error(err))
	}
}
