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
	"cli-template/internal/service1/webapi"

	_ "cli-template/api/service1/docs" // swagger 一定要有这行,指向你的文档地址

	"github.com/xiaohangshuhub/go-workit/pkg/workit"
)

func main() {

	builder := workit.NewWebAppBuilder()

	builder.AddConfig(func(build workit.ConfigBuilder) {
		build.AddYamlFile("./application.yaml")
	})

	app := builder.Build()

	app.UseRecovery()

	app.UseLogger()

	if app.Environment().IsDevelopment {
		app.UseSwagger()
	}

	app.UseAuthorization()

	app.MapRoutes(webapi.Hello)

	app.Run()
}
