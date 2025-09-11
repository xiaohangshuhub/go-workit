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
	"github.com/xiaohangshuhub/go-workit/pkg/app"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp"
)

func main() {
	// web应用构建器
	builder := webapp.NewBuilder()

	// 配置构建器(注册即生效)
	builder.AddConfig(func(build app.ConfigBuilder) {
		build.AddYamlFile("./application.yaml")
		build.AddJsonFile("./application.json")
	})

	// 构建Web应用
	app := builder.Build()

	// swag
	if app.Env().IsDevelopment {
		app.UseSwagger()
	}

	// 配置路由
	app.MapRouter(webapi.Hello)

	// 运行应用
	app.Run()
}
