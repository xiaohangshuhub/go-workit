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
	"github.com/gin-gonic/gin"
	_ "github.com/xiaohangshu-dev/go-workit/api/service1/docs" // swagger 一定要有这行,指向你的文档地址
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp"
)

func main() {
	builder := webapp.NewBuilder()

	// 构建Web应用
	// 配置 use_default_logger: false   use_default_recover: false
	// 添加默认中间件(recovery, logger),也可以通过UseMiddleware添加自定义恢复和日志中间件
	app := builder.Build().UseRecovery().UseLogger()

	app.MapRoute(func(router *gin.Engine) {
		router.GET("/hello", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Hello, World!",
			})
		})
	})

	// 运行应用
	app.Run()
}
