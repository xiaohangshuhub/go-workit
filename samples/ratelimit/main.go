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
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/xiaohangshuhub/go-workit/api/service1/docs" // swagger 一定要有这行,指向你的文档地址
	"github.com/xiaohangshuhub/go-workit/pkg/webapp"
	"github.com/xiaohangshuhub/go-workit/pkg/webapp/ratelimit"
)

func main() {
	// web应用构建器
	builder := webapp.NewBuilder()

	builder.AddRateLimiter(func(options *ratelimit.Options) {
		options.DefaultPolicy = "default"
		options.AddFixedWindowLimiter("default", func(opts *ratelimit.FixedWindowOptions) {
			opts.PermitLimit = 1                              // 每时间窗口允许的请求数
			opts.Window = time.Minute                         // 时间窗口长度
			opts.QueueProcessingOrder = ratelimit.OldestFirst // 可选，处理排队顺序
		})

	})

	// 构建Web应用
	app := builder.Build()

	app.UseRateLimiter()

	// 配置路由
	app.MapRouter(func(router *gin.Engine) {
		router.GET("/hello", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Hello, World!",
			})
		})
	})

	// 运行应用
	app.Run()
}
