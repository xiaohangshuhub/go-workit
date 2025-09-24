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
	"github.com/xiaohangshuhub/go-workit/internal/service1/webapi"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/auth"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/auth/scheme/jwt"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/authz"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/router"
)

func main() {

	builder := webapp.NewBuilder()

	builder.AddRouter(func(options *router.Options) {

		// 路由组
		hello := options.MapGroup("api/v1/hello").WithAllowAnonymous()
		hello.MapPost("", webapi.HelloNewb).WithRateLimiter("limiter")
		hello.MapGet("", webapi.HelloNewb).WithRateLimiter("limiter")

		// 路由
		options.MapGet("/hello2", webapi.HelloNewb).WithAuthenticationScheme("jwt")

		// 只注册配置,不添加路由可以搭配 MapRouter 使用
		options.MapGet("/hello", nil).WithAllowAnonymous()

		options.MapGet("hello3", webapi.HelloNewb).WithAuthorizationPolicy("admin")

	})

	builder.AddAuthentication(func(options *auth.Options) {

		options.DefaultScheme = "local_jwt_bearer"

		options.AddJwtBearer("local_jwt_bearer", func(options *jwt.Options) {

			options.TokenValidationParameters = jwt.TokenValidationParameters{
				ValidateIssuer:           true,
				ValidateAudience:         true,
				ValidateLifetime:         true,
				ValidateIssuerSigningKey: true,
				SigningKey:               []byte("secret"),
				ValidIssuer:              "sample",
				ValidAudience:            "sample",
				RequireExpiration:        true,
			}
		})
	})

	builder.AddAuthorization(func(options *authz.Options) {

		options.DefaultPolicy = "admin_role_policy"

		options.RequireRole("admin", "admin")

	})

	app := builder.Build()

	if app.Env().IsDevelopment {
		app.UseSwagger()
	}

	app.UseAuthentication()

	app.UseAuthorization()

	app.MapRoute(webapi.Hello)

	app.UseRouting()

	app.Run()
}
