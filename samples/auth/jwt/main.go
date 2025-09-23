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
)

func main() {
	builder := webapp.NewBuilder()

	builder.AddServices()

	builder.AddAuthentication(func(options *auth.Options) {
		options.DefaultScheme = "local_jwt_bearer"
		options.AddJwtBearer("local_jwt_bearer", func(options *jwt.Options) {
			options.TokenValidationParameters = jwt.TokenValidationParameters{
				ValidateIssuer:           true,
				ValidateAudience:         true,
				ValidateLifetime:         true,
				ValidateIssuerSigningKey: true,
				SigningKey:               []byte("Sinsegye Automation IDE&UI Group YYDS"),
				ValidIssuer:              "Sinsegye SF8010",
				ValidAudience:            "Sinsegye SF8010 User",
				RequireExpiration:        true,
			}
		})

	})

	builder.AddAuthorization(func(options *authz.Options) {

		options.DefaultPolicy = "admin_role_policy"

	})

	app := builder.Build()

	app.UseRecovery()

	app.UseLogger()

	if app.Env().IsDevelopment {
		app.UseSwagger()
	}
	app.UseAuthentication()

	app.UseAuthorization()

	app.MapRouter(webapi.Hello)

	app.Run()
}
