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
)

func main() {
	// web应用构建器
	builder := workit.NewWebAppBuilder()

	// 配置构建器(注册即生效)
	builder.AddConfig(func(build workit.ConfigBuilder) { build.AddYamlFile("./application.yaml") })

	// 注册服务
	builder.AddServices(fx.Provide(config.NewApplicationConfig))

	//注册鉴权方案
	builder.AddAuthentication(func(options *workit.AuthenticateOptions) {

		options.DefaultScheme = "local_jwt_bearer"

	}).
		//	本地jwt_bearer方案
		AddJwtBearer("local_jwt_bearer", func(options *workit.JwtBearerOptions) {

			options.TokenValidationParameters = workit.TokenValidationParameters{
				ValidateIssuer:           true,
				ValidateAudience:         true,
				ValidateLifetime:         true,
				ValidateIssuerSigningKey: true,
				SigningKey:               []byte("secret"),
				ValidIssuer:              "sample",
				ValidAudience:            "sample",
				RequireExpiration:        true,
			}
		}).

		//	oauth2 jwt_bearer方案
		AddJwtBearer("oauth2_jwt_bearer", func(options *workit.JwtBearerOptions) {

			options.Authority = "http://localhost:8090"
			options.RequireHttpsMetadata = false
			options.TokenValidationParameters = workit.TokenValidationParameters{
				ValidateIssuer: true,
				ValidIssuer:    "http://localhost:8090",
			}

		})

	// 注册授权策略
	builder.AddAuthorization(func(options *workit.AuthorizeOptions) {

		options.DefaultPolicy = "admin_role_policy"

	}).RequireRolePolicy("admin_role_policy", "admin", "super_admin")

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
	app.MapRoutes(webapi.Hello)

	// 配置grpc服务
	app.MapGrpcServices(hello.NewHelloService)

	// 运行应用
	app.Run()
}
