// Package workit 是一个轻量级的 Go Web 框架。
//
// workit 提供依赖注入、鉴权授权、Swagger 文档生成、配置管理、中间件支持等功能。
// 适用于快速构建 RESTful API 和 Web 服务。
//
// 功能概览：
//
//   - 依赖注入：管理服务生命周期和组件依赖。
//   - 鉴权授权：支持路由级别鉴权方案和基于角色的访问控制。
//   - Swagger 集成：自动生成 OpenAPI/Swagger 文档。
//   - 配置管理：支持 YAML、JSON、环境变量、命令行参数等多种配置格式。
//   - 中间件支持：方便添加日志、错误处理、限流等中间件。
//
// 使用示例：
//
// package main

// import (
// 	_ "github.com/xiaohangshuhub/go-workit/api/service1/docs" // swagger 一定要有这行,指向你的文档地址
// 	"github.com/xiaohangshuhub/go-workit/config"
// 	"github.com/xiaohangshuhub/go-workit/internal/service1/grpcapi/hello"
// 	"github.com/xiaohangshuhub/go-workit/internal/service1/webapi"
// 	"github.com/xiaohangshuhub/go-workit/pkg/workit"
// 	"go.uber.org/fx"
// 	"go.uber.org/zap"
// )

// func main() {
// 	// web应用构建器
// 	builder := workit.NewWebAppBuilder()

// 	// 配置构建器(注册即生效)
// 	builder.AddConfig(func(build workit.ConfigBuilder) { build.AddYamlFile("./application.yaml") })

// 	// 注册服务
// 	builder.AddServices(fx.Provide(func(logger *zap.Logger) {}))

// 	//注册鉴权方案
// 	builder.AddAuthentication().AddJwtBearer(func(options *workit.JwtBearerOptions) {})

// 	// 注册授权策略
// 	builder.AddAuthorization(config.Authorize...).RequireRole("admin_role_policy", "admin", "super_admin")

// 	// 构建Web应用
// 	app, err := builder.Build()

// 	if err != nil {
// 		app.Logger().Error("Failed to build application: %v\n", zap.Error(err))
// 		return
// 	}

// 	if app.Env().IsDevelopment {
// 		app.UseSwagger()
// 	}

// 	// 配置鉴权
// 	app.UseAuthentication()

// 	// 配置授权
// 	app.UseAuthorization()

// 	// 配置路由
// 	app.MapRoutes(webapi.Hello)

// 	// 配置grpc服务
// 	app.MapGrpcServices(hello.NewHelloService)

// 	// 运行应用
// 	if err := app.Run(); err != nil {
// 		app.Logger().Error("Error running application", zap.Error(err))
// 	}
// }

package workit
