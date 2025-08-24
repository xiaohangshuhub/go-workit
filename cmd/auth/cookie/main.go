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
	"go.uber.org/zap"
)

func main() {
	// web应用构建器
	builder := workit.NewWebAppBuilder()

	// 配置构建器(注册即生效)
	builder.AddConfig(func(build workit.ConfigBuilder) { build.AddYamlFile("./application.yaml") })

	// 注册服务
	builder.AddServices()

	//注册鉴权方案
	builder.AddAuthentication(func(options *workit.AuthenticateOptions) {

		options.DefaultScheme = "local_jwt_bearer"
		options.UseAllowAnonymous(config.SkipRoutes...)
		options.UseRouteSchemes(config.RouteAuthenticationSchemes...)

	}).
		//	本地jwt_bearer方案
		AddCookie("cookie_auth", func(options *workit.CookieOptions) {

			options.Name = "auth_token"
			options.DataProtectionKey = "my_secret_key"
		})

	// 注册授权策略
	builder.AddAuthorization(func(options *workit.AuthorizeOptions) {

		options.DefaultPolicy = "admin_role_policy"
		options.UseRoutePolicies(config.Authorize...)

	}).RequireRolePolicy("admin_role_policy", "admin", "super_admin")

	// 构建Web应用
	app, err := builder.Build()

	if err != nil {
		app.Logger().Error("Failed to build application: %v\n", zap.Error(err))
		return
	}

	if app.Env().IsDevelopment {
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
		app.Logger().Error("Error running application", zap.Error(err))
	}
}
