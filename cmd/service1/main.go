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
	"fmt"

	_ "cli-template/api/service1/docs" // swagger 一定要有这行,指向你的文档地址

	"github.com/xiaohangshuhub/go-workit/pkg/workit"
	"go.uber.org/zap"
)

func main() {

	builder := workit.NewWebAppBuilder()

	builder.AddConfig(func(build workit.ConfigBuilder) {
		build.AddYamlFile("./config.yaml")
	})

	app, err := builder.Build()

	if err != nil {
		fmt.Printf("Failed to build application: %v\n", err)
		return
	}

	if app.Env().IsDevelopment {
		app.UseSwagger()
	}

	app.UseAuthorization()

	app.MapRoutes(webapi.Hello)

	if err := app.Run(); err != nil {
		app.Logger().Error("Error running application", zap.Error(err))
	}
}
