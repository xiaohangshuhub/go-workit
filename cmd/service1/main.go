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
	_ "cli-template/api/service1/docs" // swagger 一定要有这行,指向你的文档地址
	"cli-template/internal/service1/webapi"

	"github.com/xiaohangshuhub/go-workit/pkg/webapp"
)

func main() {

	builder := webapp.NewBuilder()

	app := builder.Build()

	if app.Env().IsDevelopment {
		app.UseSwagger()
	}

	app.MapRouter(webapi.Hello)

	app.Run()
}
