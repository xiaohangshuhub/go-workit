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
	"github.com/labstack/echo/v4"
	_ "github.com/xiaohangshuhub/go-workit/api/service1/docs" // swagger 一定要有这行,指向你的文档地址
	"github.com/xiaohangshuhub/go-workit/pkg/webapp"
	echoapp "github.com/xiaohangshuhub/go-workit/pkg/webapp/echo"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/web"
)

func main() {
	// web应用构建器
	builder := webapp.NewBuilder()

	// 构建Web应用
	app := builder.Build(func(b *webapp.WebApplicationBuilder) web.Application {

		return echoapp.NewEchoWebApplication(web.InstanceConfig{
			Config:     b.Config,
			Logger:     b.Logger,
			Container:  b.Container,
			Applicaton: b.Application,
			Router:     b.Router,
		})

	})

	// 配置路由
	app.MapRouter(func(router *echo.Echo) {

		router.GET("/hello", func(c echo.Context) error {

			return c.JSON(200, map[string]string{
				"message": "Hello, ECHO!",
			})
		})
	})

	// 运行应用
	app.Run()
}
