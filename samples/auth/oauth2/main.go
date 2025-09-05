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
	"github.com/xiaohangshuhub/go-workit/internal/service1/grpcapi/hello"
	"github.com/xiaohangshuhub/go-workit/internal/service1/webapi"
	"github.com/xiaohangshuhub/go-workit/pkg/workit"
)

func main() {
	// web应用构建器
	builder := workit.NewWebAppBuilder()

	// 配置构建器(注册即生效)
	builder.AddConfig(func(build workit.ConfigBuilder) { build.AddYamlFile("./application.yaml") })

	// 注册服务
	builder.AddServices()

	//注册鉴权方案
	builder.AddAuthentication(func(options *workit.AuthenticationOptions) {

		options.DefaultScheme = "oauth2_jwt_bearer"

	}).AddJwtBearer("oauth2_jwt_bearer", func(options *workit.JwtBearerOptions) {

		options.Authority = "http://localhost:8090"
		options.RequireHttpsMetadata = false
		options.TokenValidationParameters = workit.TokenValidationParameters{
			ValidateIssuer: true,
			ValidIssuer:    "http://localhost:8090",
		}

	})

	// 注册授权策略
	builder.AddAuthorization(func(options *workit.AuthorizationOptions) {

		options.DefaultPolicy = "admin_role_policy"

	}).RequireRole("admin_role_policy", "admin", "super_admin")

	// 构建Web应用
	app := builder.Build()

	if app.Env().IsDevelopment {
		app.UseSwagger()
	}

	// 配置鉴权
	app.UseAuthentication()

	// 配置授权
	app.UseAuthorization()

	// 配置路由
	app.MapRouter(webapi.Hello)

	// 配置grpc服务
	app.MapGrpcServices(hello.NewHelloService)

	// 运行应用
	app.Run()
}
